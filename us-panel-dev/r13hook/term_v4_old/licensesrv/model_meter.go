package licensesrv

import (
	"errors"
	"time"
)

// Hourly balance is integer wall seconds. Reserve up to the request deadline;
// successful calls charge ceil(upstream wall duration), including cold start,
// excluding admission queue time. Failure/cancel releases the entire hold.
func (s *Store) reserveModel(key string, budget time.Duration, now time.Time) (string, int, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := randID()
	amount := 1
	hourly := false
	err := s.transactionLocked(func() error {
		s.expireLocked(now)
		var e error
		key, e = s.modelLicenseLocked(key, now)
		if e != nil {
			return e
		}
		l, ok := s.st.Licenses[key]
		if !ok || l.Revoked || !now.Before(l.ExpiresAt) {
			return ErrInvalid
		}
		if l.Plan.Kind == "use" {
			hourly = true
			if l.MeterUnit == "" {
				if l.Credits != 0 || l.Reserved != 0 {
					return errors.New("hourly legacy balance needs review")
				}
				for _, v := range s.st.Settlements {
					if v.Key == key {
						return errors.New("hourly legacy usage needs review")
					}
				}
				l.Credits = l.Plan.Seconds
				l.MeterUnit = "seconds"
			}
			amount = int(budget / time.Second)
			if amount < 1 {
				amount = 1
			}
			if available := l.Credits - l.Reserved; amount > available {
				amount = available
			}
		} else if l.Plan.Kind != "sub" {
			return ErrInvalid
		}
		if amount <= 0 || l.Credits-l.Reserved < amount {
			return ErrNoCredit
		}
		l.Reserved += amount
		s.st.Reservations[id] = &Reservation{CreatedAt: now, Key: key, Amount: amount, Expires: now.Add(budget + 5*time.Minute)}
		return nil
	})
	return id, amount, hourly, err
}
func (s *Store) settleAmount(key, id string, amount int, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if done, ok := s.st.Settlements[id]; ok {
		if done.Key == key && done.Amount == amount && done.Commit == (amount > 0) {
			return nil
		}
		return ErrInvalid
	}
	r, ok := s.st.Reservations[id]
	if !ok || r.Key != key || amount < 0 || amount > r.Amount {
		return ErrInvalid
	}
	return s.transactionLocked(func() error {
		l, ok := s.st.Licenses[key]
		if !ok {
			return ErrInvalid
		}
		if amount > 0 && (l.Revoked || !now.Before(l.ExpiresAt) || !now.Before(s.ReservationLockedUntil(r))) { // R13: shared TTL+grace deadline
			return ErrInvalid
		}
		l.Reserved -= r.Amount
		l.Credits -= amount
		s.st.Settlements[id] = Settlement{CreatedAt: r.CreatedAt, Key: key, Commit: amount > 0, Amount: amount, SettledAt: now}
		delete(s.st.Reservations, id)
		return nil
	})
}
