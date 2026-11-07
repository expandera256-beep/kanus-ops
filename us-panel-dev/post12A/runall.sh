cd /workspace/us-panel-dev/post12A/term
go test -count=1 -json ./... > ../after_test.jsonl 2>&1; echo "test_exit=$?" > ../after_exit.txt
go test -race -count=1 ./... > ../after_race.log 2>&1; echo "race_exit=$?" >> ../after_exit.txt
cd ../orig; go test -count=1 -json ./... > ../before_test.jsonl 2>&1; echo "before_exit=$?" >> ../after_exit.txt
echo DONE >> ../after_exit.txt
