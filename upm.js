(function(){
if(window.__mbUPM)return;
var U=window.__mbUPM={state:'active',_last:Date.now(),_total:0};
U._reset=function(){U.state='active';U._last=Date.now();};
['mousemove','mousedown','keydown','touchstart','scroll','click'].forEach(function(e){
document.addEventListener(e,function(){U._reset();},{passive:true});});
setInterval(function(){
var idle=(Date.now()-U._last)/1000;
if(idle>30){U.state='idle';U._total++;}
if(idle>120){U.state='away';}
},5000);
})();
