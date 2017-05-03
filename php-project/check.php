<?php

$post_data = $_POST;
$header = get_all_headers();

$sessionid = $header['sessionid'];

$host = '127.0.0.1';
$port = '6379';
$timeout = 0;

$redis = new Redis();
$redis->connect($host, $port, $timeout);
$session_content = $redis->get("miniappsession:".$sessionid);
if($session_content)
{
    echo $session_content;
} else {
    echo 0;
}

/**
 * 获取自定义的header数据
 */
function get_all_headers(){

    // 忽略获取的header数据
    $ignore = array('host','accept','content-length','content-type');

    $headers = array();

    foreach($_SERVER as $key=>$value){
        if(substr($key, 0, 5)==='HTTP_'){
            $key = substr($key, 5);
            $key = str_replace('_', ' ', $key);
            $key = str_replace(' ', '-', $key);
            $key = strtolower($key);

            if(!in_array($key, $ignore)){
                $headers[$key] = $value;
            }
        }
    }

    return $headers;

}