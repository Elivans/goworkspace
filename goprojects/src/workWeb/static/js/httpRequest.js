
function CreateXMLHttpRequest(){
	var xRequest=null;
  	if (typeof ActiveXObject != "undefined"){
    	//Internet Explorer     
    	xRequest=new ActiveXObject("Microsoft.XMLHTTP");   
  	}
  	else if (window.XMLHttpRequest) {                      
		//Mozilla/Safari
    	xRequest=new XMLHttpRequest();   
	}

  	return xRequest;
}
function PostInfo(url,parm) //发送POST请求
{
	var xmlobj = CreateXMLHttpRequest(); //创建对象 
	xmlobj.open("POST", url, false); //调用
	xmlobj.setRequestHeader("Content-Type","multipart/form-data");
	
	xmlobj.send(parm); //设置为发送给服务器数据
	return xmlobj.responseText;
}