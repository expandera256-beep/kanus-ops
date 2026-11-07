<?php
// kanus.org günlük kur anlık görüntüsü (Hostinger cron, hafta içi 16:05 TSİ). PHP >= 8.0 (Hostinger sürümü doğrulanmadı; 8.x varsayıldı, test PHP 8.3.31 CLI).
// Yer (Seçenek A): public_html/_fx/fx_cron.php (+ .htaccess tümünü reddeder). Yalnız CLI (cron) çalışır;
// HTTP ile çağrılırsa 403 döner ve hiçbir şey yazmaz. Yazdığı TEK hedef: public_html/fx.json
// (geçici dosya _fx/ içinde, aynı dosya sisteminde rename). Kilit: _fx/fx_cron.lock.
// Çıkış: 0 yazıldı, 2 geçersiz/eski/ağ hatası (fx.json'a dokunulmaz), 3 başka kopya çalışıyor.
// Test için yalnız kaynak ve zaman aşımı değiştirilebilir (KANUS_FX_SRC, KANUS_FX_TIMEOUT); hedef sabit.
declare(strict_types=1);
if (PHP_SAPI !== 'cli') { if (!headers_sent()) { http_response_code(403); header('Content-Type: text/plain'); } echo "Forbidden\n"; exit; }
const FX_URL = 'https://www.tcmb.gov.tr/kurlar/today.xml';
const MAX_AGE_DAYS = 5;
set_time_limit(90);
date_default_timezone_set('Europe/Istanbul');
function fail(int $code, string $why): void { fwrite(STDERR, "fx_cron: $why\n"); exit($code); }

$dir = __DIR__;
$pub = realpath($dir . '/..');
if ($pub === false || !is_dir($pub) || basename($dir) !== '_fx') fail(2, 'must run from public_html/_fx/');
$out = $pub . '/fx.json';

$lock = fopen($dir . '/fx_cron.lock', 'c');
if ($lock === false) fail(2, 'lock open failed');
if (!flock($lock, LOCK_EX | LOCK_NB)) fail(3, 'already running');

$src = getenv('KANUS_FX_SRC') ?: FX_URL;
$timeout = max(1, min(20, (int)(getenv('KANUS_FX_TIMEOUT') ?: 20)));
$ctx = stream_context_create(['http' => ['timeout' => $timeout, 'user_agent' => 'kanus-fx/1', 'follow_location' => 0, 'ignore_errors' => false]]);
$raw = @file_get_contents($src, false, $ctx, 0, 524289);
if ($raw === false || $raw === '') fail(2, 'fetch failed');
if (strlen($raw) > 524288) fail(2, 'too large');
if (stripos($raw, '<!DOCTYPE') !== false || stripos($raw, '<!ENTITY') !== false) fail(2, 'doctype/entity');
libxml_use_internal_errors(true);
$x = simplexml_load_string($raw, 'SimpleXMLElement', LIBXML_NONET);
if ($x === false || $x->getName() !== 'Tarih_Date') fail(2, 'bad xml');
if (!preg_match('/^(\d\d)\.(\d\d)\.(\d{4})$/', (string)$x['Tarih'], $m) || !checkdate((int)$m[2], (int)$m[1], (int)$m[3])) fail(2, 'bad Tarih');
$date = "$m[3]-$m[2]-$m[1]";
$age = (int)(new DateTimeImmutable($date))->diff(new DateTimeImmutable('today'))->format('%r%a');
if ($age < -1 || $age > MAX_AGE_DAYS) fail(2, "stale/future ($age days)");
$r = [];
foreach ($x->Currency as $c) {
    $k = (string)$c['CurrencyCode'];
    if ($k !== 'USD' && $k !== 'EUR') continue;
    $s = trim((string)$c->ForexSelling);
    if (trim((string)$c->Unit) !== '1' || !preg_match('/^\d{1,4}\.\d{1,6}$/', $s)) fail(2, "bad $k");
    $v = (float)$s;
    if ($v < 1 || $v > 1000) fail(2, "range $k");
    $r[$k] = $v;
}
if (!isset($r['USD'], $r['EUR'])) fail(2, 'missing rate');
$cross = $r['EUR'] / $r['USD'];
if ($cross < 0.8 || $cross > 1.6) fail(2, 'cross');
$d = ['schema' => 'kanus.fx/1', 'source' => 'TCMB today.xml ForexSelling (döviz satış)', 'source_url' => FX_URL,
    'tcmb_tarih' => (string)$x['Tarih'], 'bulten_no' => (string)$x['Bulten_No'], 'date' => $date, 'base' => 'TRY',
    'rates' => ['USD' => $r['USD'], 'EUR' => $r['EUR']], 'generated_at' => gmdate('Y-m-d\TH:i:sP')];
$json = json_encode($d, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES | JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR) . "\n";
$tmp = tempnam($dir, 'fx.json.tmp.');
if ($tmp === false || dirname($tmp) !== $dir) { if ($tmp) @unlink($tmp); fail(2, 'tempnam'); }
if (file_put_contents($tmp, $json) !== strlen($json) || !chmod($tmp, 0644) || !rename($tmp, $out)) { @unlink($tmp); fail(2, 'write failed'); }
echo "fx_cron: ok {$d['tcmb_tarih']} USD={$r['USD']} EUR={$r['EUR']}\n";
