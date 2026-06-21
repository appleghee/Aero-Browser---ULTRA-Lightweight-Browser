(function(){
if(window.__mbAVP)return;
var A=window.__mbAVP={
velocity:0,direction:0,samples:[],_pred:0,_hits:0,
lastY:0,lastTime:0,
 init:function(){
 var T=this;
 T.lastY=window.scrollY;
 T.lastTime=Date.now();
 T._predictThrottled=0;
 document.addEventListener('scroll',function(){
 var now=Date.now(),sy=window.scrollY;
 var dt=Math.max(16,now-T.lastTime);
 var vel=Math.abs(sy-T.lastY)/dt*1000;
 T.velocity=Math.min(5000,T.velocity*0.7+vel*0.3);
 T.direction=sy>T.lastY?1:-1;
 T.lastY=sy;T.lastTime=now;
 if(now-T._predictThrottled>200){T._predictThrottled=now;T._predict();}
 },{passive:true});
 setInterval(function(){T.velocity*=0.85;},1000);
 },
 _predict:function(){
 var T=this;
 var ahead=T.velocity*0.3;
 if(ahead<50)return;
 T._pred++;
 var sel='article,section';
 var els=document.querySelectorAll(sel);
 var best=null,bestDist=Infinity;
 for(var i=0;i<els.length&&i<8;i++){
 var el=els[i],r=el.getBoundingClientRect();
 var dist=Math.abs(r.top-ahead*T.direction);
 if(dist<bestDist&&dist>0){bestDist=dist;best=el;}}
 if(best&&bestDist<window.innerHeight*0.5){
 var imgs=best.querySelectorAll('img[loading=lazy]');
 for(var j=0;j<imgs.length;j++){imgs[j].loading='eager';}
 T._hits++;
 }
 }
};
if(document.body){A.init();}else{document.addEventListener('DOMContentLoaded',function(){A.init();})}
})();
