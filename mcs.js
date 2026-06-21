(function(){
if(window.__mbMCS)return;
var M=window.__mbMCS={_def:0,_exe:0,_queue:[],_running:false,_cancel:{}};
M._tick=function(){
if(M._queue.length===0||M._running){M._running=false;return;}
M._running=true;
var task=M._queue.shift();
M._exe++;
try{if(!task.cancelled)task.fn();}catch(e){}
setTimeout(function(){M._tick();},5);
};
M.defer=function(fn,id){
M._def++;
var entry={fn:fn,cancelled:false};
if(id)M._cancel[id]=entry;
M._queue.push(entry);
if(!M._running)M._tick();
};
var _st=setTimeout;
var _ct=clearTimeout;
setTimeout=function(cb,ms){
if(typeof cb==='function'&&(!ms||ms>50)){
var id=_st(function(){},ms);
M.defer(cb,id);
return id;
}
return _st(cb,ms);
};
clearTimeout=function(id){
_ct(id);
if(M._cancel[id]){M._cancel[id].cancelled=true;delete M._cancel[id];}
};
})();