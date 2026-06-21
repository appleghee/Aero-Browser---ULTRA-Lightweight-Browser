(function(){
if(window.__mbHFS)return;
var H=window.__mbHFS={_hot:0,_store:{}};
H.access=function(path){
var now=Date.now();
if(!H._store[path]){
H._store[path]={hits:0,lastSeen:now};
}
var e=H._store[path];
e.hits++;
e.lastSeen=now;
if(e.hits>3)H._hot++;
};
H.cool=function(){
var now=Date.now();
Object.keys(H._store).forEach(function(k){
var e=H._store[k];
if(now-e.lastSeen>60000){
e.hits=Math.max(0,e.hits-1);
}
});
};
var _fi=window.fetch;
window.fetch=function(u,o){
var url=typeof u==='string'?u:u&&u.url?u.url:'';
if(url)H.access(url.split('?')[0]);
return _fi.call(this,u,o);
};
setInterval(function(){H.cool();},30000);
})();
