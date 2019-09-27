// 1.判断select选项中 是否存在Value="paraValue"的Item 
function jsSelectIsExitItem(objSelect, objItemValue) { 
  var isExit = false; 
  for (var i = 0; i < objSelect.options.length; i++) { 
    if (objSelect.options[i].value == objItemValue) { 
      isExit = true; 
      break; 
    } 
  } 
  return isExit; 
} 

// 2.向select选项中 加入一个Item 
function jsAddItemToSelect(objSelect, objItemText, objItemValue) { 
  //判断是否存在 
  if (jsSelectIsExitItem(objSelect, objItemValue)) { 
    alert("该Item的Value值已经存在"); 
  } else { 
    var varItem = new Option(objItemText, objItemValue); 
   objSelect.options.add(varItem); 
   alert("成功加入"); 
  } 
} 

 /*
     第一步 获取name属性为luck值得对象数组(节点数组)
 */
var paramElements = document.getElementsByName("maywideParams");

 /*
     第二步 遍历节点数组
 */
for(var i=0;i<paramElements.length;i++){
  //获取元素的value值
  alert(paramElements[i].value);
  //获取元素的type值
  alert(paramElements[i].type);
  alert(paramElements[i].class);
     
	
}