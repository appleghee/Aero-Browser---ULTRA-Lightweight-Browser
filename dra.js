(function(){
if(window.__mbDRA)return;
var D=window.__mbDRA={_throttled:0,_budget:100};
D.adjust=function(pct){D._budget=Math.max(20,Math.min(100,pct));};
var _fi=window.fetch;
window.fetch=function(u,o){
if(D._budget<50&&Math.random()>D._budget/100){
D._throttled++;
return Promise.resolve(new Response('',{status:429,statusText:'DRA throttled'}));
}
return _fi.call(this,u,o);
};
setInterval(function(){
var mem=performance.memory;
if(mem){
var ratio=mem.usedJSHeapSize/mem.jsHeapSizeLimit;
D._budget=100-Math.round(ratio*60);
}
},5000);
})();
