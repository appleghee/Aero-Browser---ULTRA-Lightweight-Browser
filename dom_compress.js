(function(){
if(window.__mbDOMC)return;
var C=window.__mbDOMC={_lastSize:0,_comp:0,_nodes:0};
var _orig=window.__mbSnapshot||null;
C.compress=function(){
var all=document.querySelectorAll('body *');
var total=all.length;
var tags={};
all.forEach(function(el,i){
if(i>=500)return;
var t=el.tagName.toLowerCase();
if(!tags[t])tags[t]=0;
tags[t]++;});
var json=JSON.stringify({tags:tags,count:total,url:location.href,title:document.title});
C._lastSize=new Blob([document.documentElement.outerHTML]).size;
C._comp=json.length;
C._nodes=total;
window.__mbCompressedSnapshot={tags:tags,count:total,url:location.href,title:document.title};
return C._nodes;
};
if(document.body)C.compress();else document.addEventListener('DOMContentLoaded',function(){C.compress();});
})();
