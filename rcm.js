(function(){
if(window.__mbRCM)return;
var R=window.__mbRCM={_models:{},_thresh:'auto'};
R.cost=function(url,type,size){
var domain=(url.split('/')[2]||'unknown').replace('www.','');
if(!R._models[domain]){
R._models[domain]={totalSize:0,totalReqs:0,avgCost:0};
}
var m=R._models[domain];
m.totalSize+=size||0;
m.totalReqs++;
m.avgCost=m.totalSize/m.totalReqs;
};
R.shouldLoad=function(url,priority){
var domain=(url.split('/')[2]||'unknown').replace('www.','');
var m=R._models[domain];
if(!m)return true;
if(priority==='high')return true;
if(m.avgCost>50000)return false;
return true;
};
var _fi=window.fetch;
window.fetch=function(u,o){
var url=typeof u==='string'?u:u&&u.url?u.url:'';
var prio=(o&&o.priority)||'auto';
if(url&&!R.shouldLoad(url,prio)){
return Promise.resolve(new Response('',{status:204}));
}
var p=_fi.call(this,u,o).then(function(r){
var cl=r.headers.get('content-length');
R.cost(url,'fetch',cl?parseInt(cl):0);
return r;
});
return p;
};
})();
