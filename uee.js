(function(){
if(window.__mbUEE)return;
var U=window.__mbUEE={_del:0,_sav:0};
U._delegates={};
U.on=function(sel,ev,fn){
if(!U._delegates[ev])U._delegates[ev]={};
U._delegates[ev][sel]=fn;
U._del++;
};
U._handle=function(e){
var ev=e.type;
if(!U._delegates[ev])return;
var target=e.target;
Object.keys(U._delegates[ev]).forEach(function(sel){
var el=target.closest(sel);
if(el){
U._sav++;
U._delegates[ev][sel].call(el,e);
}
});
};
['click','mousedown','mouseup','keydown','keyup','change'].forEach(function(ev){
document.addEventListener(ev,function(e){U._handle(e);},true);
});
var _ael=EventTarget.prototype.addEventListener;
EventTarget.prototype.addEventListener=function(ev,fn,opts){
if(ev==='click'||ev==='mousedown'||ev==='keydown'){
U._del++;
return _ael.call(this,ev,fn,opts);
}
return _ael.call(this,ev,fn,opts);
};
})();
