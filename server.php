<?php
error_reporting(0);

if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    header('Access-Control-Allow-Origin: *');
    header('Access-Control-Allow-Methods: GET, POST, PUT, DELETE');
    header('Access-Control-Allow-Headers: Content-Type');
    http_response_code(204);
    exit;
} else {
	header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: GET, POST, PUT, DELETE,OPTIONS');
header('Access-Control-Allow-Headers: Content-Type');
}

$c=file_get_contents("php://input");

$date=date("Y-m-d H:i:s");
$data=json_encode(["code"=>0,"data"=>$_SERVER,"date"=>$date]);
file_put_contents('http_result.txt',$data,FILE_APPEND);

echo($data);