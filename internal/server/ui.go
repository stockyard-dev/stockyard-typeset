package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html>
<html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Typeset</title>
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#c45d2c;--rl:#e8753a;--leather:#a0845c;--ll:#c4a87a;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c44040;--mono:'JetBrains Mono',Consolas,monospace;--serif:'Libre Baskerville',Georgia,serif}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);font-size:13px;line-height:1.6;height:100vh;overflow:hidden}
a{color:var(--rl);text-decoration:none}a:hover{color:var(--gold)}
.app{display:flex;height:100vh}
.sidebar{width:230px;background:var(--bg2);border-right:1px solid var(--bg3);display:flex;flex-direction:column;flex-shrink:0;overflow-y:auto}
.sidebar-hdr{padding:.6rem .8rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.sidebar-hdr span{font-family:var(--serif);font-size:.9rem;color:var(--rl)}
.site-select{width:100%;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.72rem;padding:.3rem .4rem;margin:.4rem .8rem;width:calc(100% - 1.6rem);outline:none}
.nav-section{padding:.3rem .8rem;font-size:.6rem;text-transform:uppercase;letter-spacing:1.5px;color:var(--rust);margin-top:.3rem;display:flex;justify-content:space-between;align-items:center}
.nav-page{padding:.25rem .8rem;padding-left:1.4rem;font-size:.73rem;cursor:pointer;color:var(--cd);transition:.1s}
.nav-page:hover{background:var(--bg3)}.nav-page.active{background:var(--bg3);color:var(--cream)}
.nav-page.draft{color:var(--cm);font-style:italic}
.content{flex:1;display:flex;flex-direction:column;min-width:0}
.content-toolbar{padding:.4rem .8rem;border-bottom:1px solid var(--bg3);display:flex;align-items:center;gap:.5rem}
.btn{font-family:var(--mono);font-size:.68rem;padding:.25rem .6rem;border:1px solid;cursor:pointer;background:transparent;transition:.15s;white-space:nowrap}
.btn-p{border-color:var(--rust);color:var(--rl)}.btn-p:hover{background:var(--rust);color:var(--cream)}
.btn-d{border-color:var(--bg3);color:var(--cm)}.btn-d:hover{border-color:var(--red);color:var(--red)}
.btn-s{border-color:var(--green);color:var(--green)}.btn-s:hover{background:var(--green);color:var(--bg)}
.page-title{width:100%;background:transparent;border:none;color:var(--cream);font-family:var(--serif);font-size:1.2rem;padding:.5rem .8rem;outline:none;border-bottom:1px solid var(--bg3)}
.editor-area{flex:1;display:flex;overflow:hidden}
.editor-area textarea{flex:1;background:transparent;border:none;color:var(--cd);font-family:var(--mono);font-size:.8rem;padding:.8rem;outline:none;resize:none;line-height:1.7}
.preview{flex:1;padding:.8rem 1.2rem;overflow-y:auto;border-left:1px solid var(--bg3);font-size:.82rem;color:var(--cd);line-height:1.8;display:none}
.preview h1,.preview h2,.preview h3{color:var(--cream);font-family:var(--serif);margin:1rem 0 .4rem}
.preview h1{font-size:1.3rem}.preview h2{font-size:1.05rem}.preview h3{font-size:.9rem}
.preview code{background:var(--bg3);padding:.1rem .3rem;font-size:.75rem}
.preview pre{background:var(--bg3);padding:.6rem;margin:.5rem 0;overflow-x:auto;font-size:.75rem}
.preview pre code{background:transparent;padding:0}
.preview ul,.preview ol{padding-left:1.2rem;margin:.4rem 0}
.preview blockquote{border-left:3px solid var(--rust);padding-left:.8rem;color:var(--cm);margin:.5rem 0}
.preview p{margin:.4rem 0}
.empty{display:flex;align-items:center;justify-content:center;flex:1;color:var(--cm);font-style:italic;font-family:var(--serif)}
.modal-bg{position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,.65);display:flex;align-items:center;justify-content:center;z-index:100}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:90%;max-width:450px}
.modal h2{font-family:var(--serif);font-size:.9rem;margin-bottom:.8rem}
label.fl{display:block;font-size:.65rem;color:var(--leather);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem;margin-top:.5rem}
input[type=text],input[type=number],select{background:var(--bg);border:1px solid var(--bg3);color:var(--cream);padding:.35rem .5rem;font-family:var(--mono);font-size:.78rem;width:100%;outline:none}
</style>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital@0;1&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
</head><body>
<div class="app">
<div class="sidebar">
<div class="sidebar-hdr"><span>Typeset</span><button class="btn btn-p" style="font-size:.55rem;padding:.1rem .3rem" onclick="showNewSite()">+ Site</button></div>
<select class="site-select" id="siteSelect" onchange="switchSite(this.value)"></select>
<div id="navTree"></div>
<div style="padding:.5rem .8rem;margin-top:.3rem">
<button class="btn btn-p" style="font-size:.6rem;width:100%" onclick="showNewPage()">+ Page</button>
<button class="btn btn-d" style="font-size:.6rem;width:100%;margin-top:.3rem" onclick="showNewSection()">+ Section</button>
</div>
<div style="margin-top:auto;padding:.5rem .8rem;border-top:1px solid var(--bg3);font-size:.6rem;color:var(--cm)" id="docsUrl"></div>
</div>
<div class="content" id="contentArea">
<div class="empty">Select or create a doc site to begin.</div>
</div>
</div>
<div id="modal"></div>
<script>
let sites=[],curSite='',curPage=null,editing=true,previewOn=false;
async function api(u,o){return(await fetch(u,o)).json()}
function esc(s){return String(s||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;')}

async function init(){
  const d=await api('/api/sites');sites=d.sites||[];
  document.getElementById('siteSelect').innerHTML='<option value="">Select site</option>'+sites.map(s=>'<option value="'+s.id+'">'+esc(s.name)+' ('+s.page_count+' pages)</option>').join('');
  if(curSite){document.getElementById('siteSelect').value=curSite;loadNav()}
}
function switchSite(id){curSite=id;curPage=null;document.getElementById('contentArea').innerHTML='<div class="empty">Select a page from the sidebar.</div>';loadNav()}

async function loadNav(){
  if(!curSite)return;
  const d=await api('/api/sites/'+curSite+'/nav');
  const nav=d.nav||[];
  const site=sites.find(s=>s.id===curSite);
  document.getElementById('docsUrl').innerHTML=site?'Public: <a href="/docs/'+esc(site.slug)+'" target="_blank">/docs/'+esc(site.slug)+'</a>':'';
  let html='';
  for(const item of nav){
    html+='<div class="nav-section">'+esc(item.section.name||'Pages');
    if(item.section.id){html+='<span style="cursor:pointer;font-size:.55rem;color:var(--cm)" onclick="delSection(\''+item.section.id+'\')">del</span>'}
    html+='</div>';
    for(const p of(item.pages||[])){
      const active=curPage&&curPage.id===p.id?'active':'';
      html+='<div class="nav-page '+active+(p.draft?' draft':'')+'" onclick="openPage(\''+p.id+'\')">'+esc(p.title)+'</div>';
    }
  }
  document.getElementById('navTree').innerHTML=html||'<div style="padding:.8rem;color:var(--cm);font-size:.72rem">No pages yet.</div>';
}

async function openPage(id){
  curPage=await api('/api/pages/'+id);editing=true;previewOn=false;renderEditor();loadNav()
}

function renderEditor(){
  if(!curPage){return}
  const p=curPage;
  document.getElementById('contentArea').innerHTML=
    '<div class="content-toolbar">'+
      '<button class="btn btn-s" onclick="savePage()">Save</button>'+
      '<button class="btn btn-d" id="prevBtn" onclick="togglePreview()">Preview</button>'+
      '<select id="pgSection" onchange="savePage()" style="background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.68rem;padding:.2rem .3rem"></select>'+
      '<label style="font-size:.6rem;color:var(--leather);display:flex;align-items:center;gap:.2rem"><input type="checkbox" id="pgDraft"'+(p.draft?' checked':'')+' onchange="savePage()"> Draft</label>'+
      '<span style="flex:1"></span>'+
      '<span style="font-size:.6rem;color:var(--cm)">'+p.word_count+'w</span>'+
      '<button class="btn btn-d" onclick="if(confirm(\'Delete?\'))delPage(\''+p.id+'\')">Del</button>'+
    '</div>'+
    '<input class="page-title" id="pgTitle" value="'+esc(p.title)+'" oninput="autoSave()">'+
    '<div class="editor-area">'+
      '<textarea id="pgBody" oninput="autoSave()">'+esc(p.body)+'</textarea>'+
      '<div class="preview" id="previewArea"></div>'+
    '</div>';
  loadSectionSelect()
}

async function loadSectionSelect(){
  if(!curSite)return;
  const d=await api('/api/sites/'+curSite+'/sections');
  const sel=document.getElementById('pgSection');
  if(sel){
    sel.innerHTML='<option value="">No section</option>'+((d.sections||[]).map(s=>'<option value="'+s.id+'"'+(curPage&&curPage.section_id===s.id?' selected':'')+'>'+esc(s.name)+'</option>').join(''));
  }
}

let saveTimer=null;
function autoSave(){if(saveTimer)clearTimeout(saveTimer);saveTimer=setTimeout(savePage,800)}

async function savePage(){
  if(!curPage)return;
  const body={title:document.getElementById('pgTitle').value||'Untitled',body:document.getElementById('pgBody').value,section_id:document.getElementById('pgSection')?document.getElementById('pgSection').value:'',draft:document.getElementById('pgDraft')?document.getElementById('pgDraft').checked:false};
  await api('/api/pages/'+curPage.id,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
  curPage=await api('/api/pages/'+curPage.id);
  if(previewOn)updatePreview();
  loadNav()
}

async function delPage(id){await api('/api/pages/'+id,{method:'DELETE'});curPage=null;document.getElementById('contentArea').innerHTML='<div class="empty">Page deleted.</div>';loadNav();init()}

function togglePreview(){
  previewOn=!previewOn;
  const pa=document.getElementById('previewArea');
  if(pa){pa.style.display=previewOn?'block':'none'}
  document.getElementById('prevBtn').textContent=previewOn?'Edit':'Preview';
  if(previewOn)updatePreview()
}

function updatePreview(){
  const md=document.getElementById('pgBody').value;
  const pa=document.getElementById('previewArea');
  if(pa)pa.innerHTML=renderMd(md)
}

function renderMd(md){
  let h=esc(md);
  h=h.replace(/^### (.+)$/gm,'<h3>$1</h3>');h=h.replace(/^## (.+)$/gm,'<h2>$1</h2>');h=h.replace(/^# (.+)$/gm,'<h1>$1</h1>');
  h=h.replace(/\*\*(.+?)\*\*/g,'<strong>$1</strong>');h=h.replace(/\*(.+?)\*/g,'<em>$1</em>');
  var bt=String.fromCharCode(96);h=h.replace(new RegExp(bt+'([^'+bt+']+)'+bt,'g'),'<code>$1</code>');
  h=h.replace(/^&gt; (.+)$/gm,'<blockquote>$1</blockquote>');
  h=h.replace(/^- (.+)$/gm,'<li>$1</li>');
  h=h.replace(/\n\n/g,'</p><p>');
  return '<p>'+h+'</p>';
}

function showNewSite(){
  document.getElementById('modal').innerHTML='<div class="modal-bg" onclick="if(event.target===this)closeModal()"><div class="modal"><h2>New Doc Site</h2><label class="fl">Name</label><input type="text" id="ns-name" placeholder="API Docs"><label class="fl">Version</label><input type="text" id="ns-ver" value="latest"><div style="display:flex;gap:.5rem;margin-top:1rem"><button class="btn btn-p" onclick="saveNewSite()">Create</button><button class="btn btn-d" onclick="closeModal()">Cancel</button></div></div></div>'
}
async function saveNewSite(){
  const b={name:document.getElementById('ns-name').value,version:document.getElementById('ns-ver').value};
  if(!b.name){alert('Name required');return}
  const r=await api('/api/sites',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(b)});
  curSite=r.id;closeModal();init()
}

function showNewSection(){
  if(!curSite){alert('Select a site first');return}
  document.getElementById('modal').innerHTML='<div class="modal-bg" onclick="if(event.target===this)closeModal()"><div class="modal"><h2>New Section</h2><label class="fl">Name</label><input type="text" id="nsec-name" placeholder="Getting Started"><label class="fl">Position</label><input type="number" id="nsec-pos" value="0"><div style="display:flex;gap:.5rem;margin-top:1rem"><button class="btn btn-p" onclick="saveNewSection()">Create</button><button class="btn btn-d" onclick="closeModal()">Cancel</button></div></div></div>'
}
async function saveNewSection(){
  const b={site_id:curSite,name:document.getElementById('nsec-name').value,position:parseInt(document.getElementById('nsec-pos').value)||0};
  if(!b.name){alert('Name required');return}
  await api('/api/sections',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(b)});closeModal();loadNav()
}
async function delSection(id){if(!confirm('Delete section?'))return;await api('/api/sections/'+id,{method:'DELETE'});loadNav()}

function showNewPage(){
  if(!curSite){alert('Select a site first');return}
  api('/api/pages',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({site_id:curSite,title:'New Page',body:'',draft:true})}).then(p=>{
    curPage=p;renderEditor();loadNav();init()
  })
}

function closeModal(){document.getElementById('modal').innerHTML=''}
init()
fetch('/api/tier').then(r=>r.json()).then(j=>{if(j.tier==='free'){var b=document.getElementById('upgrade-banner');if(b)b.style.display='block'}}).catch(()=>{var b=document.getElementById('upgrade-banner');if(b)b.style.display='block'});
</script></body></html>`
