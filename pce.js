(function(){
if(window.__mbPCE)return;
var P=window.__mbPCE={_batched:0,_flushed:0,_queue:[],_timer:null};
P._flush=function(){
if(P._queue.length===0)return;
P._flushed+=P._queue.length;
P._queue=[];
P._timer=null;
};
P.batch=function(fn){
P._queue.push(fn);
P._batched++;
if(!P._timer)P._timer=setTimeout(P._flush,50);
};
var _mo=window.MutationObserver;
window.MutationObserver=function(cb){
var _cb=function(muts){
P.batch(function(){cb(muts);});
};
var obs=new _mo(_cb);
var _ob=obs.observe.bind(obs);
obs.observe=function(t,c){P._flushed++;return _ob(t,c);};
return obs;
};
})();
