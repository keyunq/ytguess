<?php
$code = $_GET['code'];
define("APPID",'yourappid');
define("SECRET",'yoursecret');
$url = "https://api.weixin.qq.com/sns/jscode2session?appid=".APPID."&secret=".SECRET."&js_code=".$code."&grant_type=authorization_code";
$rs = curlGet($url);

$arr = json_decode($rs);

$str = randomFromDev(32);

$host = '127.0.0.1';
$port = '6379';
$timeout = 0;

$redis = new Redis();
$redis->connect($host, $port, $timeout);
$expires_time = 15*24*60*60;
$session_content = md5($arr->openid.$expires_time);
$redis->setex("miniappsession:".$str,$expires_time,$session_content);

$sessionObj['k'] = $str;
$sessionObj['v'] = $session_content;
echo json_encode($sessionObj);


function randomFromDev($len)
{
    $fp = @fopen('/dev/urandom','rb');
    $result = '';
    if ($fp !== FALSE) {
        $result .= @fread($fp, $len);
        @fclose($fp);
    }
    else
    {
        trigger_error('Can not open /dev/urandom.');
    }
    $result = md5($result);
    // convert from binary to string
    //$result = base64_encode($result);
    // remove none url chars
    //$result = strtr($result, '+/', '-_');
    // Remove = from the end
    //$result = str_replace('=', ' ', $result);
    return $result;
}

function curlGet($url, $method = 'get', $data = '')
{
    $ch = curl_init();
    $header = 'Accept-Charset: utf-8';
    curl_setopt($ch, CURLOPT_URL, $url);
    curl_setopt($ch, CURLOPT_CUSTOMREQUEST, strtoupper($method));
    curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
    curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, false);
    curl_setopt($ch, CURLOPT_HTTPHEADER, $header);
    curl_setopt($ch, CURLOPT_USERAGENT, 'Mozilla/5.0 (compatible; MSIE 5.01; Windows NT 5.0)');
    curl_setopt($ch, CURLOPT_FOLLOWLOCATION, 1);
    curl_setopt($ch, CURLOPT_AUTOREFERER, 1);
    curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    $temp = curl_exec($ch);
    return $temp;
}