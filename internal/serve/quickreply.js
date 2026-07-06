(function() {
'use strict';

const lang = (document.documentElement.lang || 'en').startsWith('zh') ? 'zh' : 'en';
const t = (zh, en) => (lang === 'zh' ? zh : en);
const AUTO_SEND_KEY = 'reasonix.quick_reply.auto_send';

const style = document.createElement('style');
style.textContent =
`.qr-shell{display:flex;flex:1;flex-direction:column;gap:14px;width:100%;min-width:0;min-height:0}
#qr-modal .modal{display:flex;flex-direction:column;width:min(760px,92vw);max-width:92vw;max-height:min(88vh,900px);overflow:hidden}
#qr-modal .modal__body{display:flex;flex:1;min-height:0;overflow:hidden}
.qr-toolbar{display:flex;align-items:center;justify-content:space-between;gap:10px;flex-wrap:wrap}
.qr-toolbar__actions{display:flex;align-items:center;gap:8px;flex-wrap:wrap}
.qr-toggle{display:inline-flex;align-items:center;gap:8px;font-size:12px;color:var(--fg-2)}
.qr-toggle input{width:auto}
.qr-tabs{display:flex;gap:2px;padding:0;flex-wrap:nowrap;overflow-x:auto;scrollbar-width:none}
.qr-tabs::-webkit-scrollbar{display:none}
.qr-tab{display:inline-flex;align-items:center;gap:4px;padding:3px 8px;border-radius:5px;font-size:11px;font-weight:500;color:var(--fg-2);cursor:pointer;white-space:nowrap;border:none;background:none;transition:background .15s,color .15s}
.qr-tab:hover{background:var(--panel);color:var(--fg)}
.qr-tab--active{background:var(--accent-soft);color:var(--accent)}
.qr-tab__badge{display:inline-flex;align-items:center;justify-content:center;min-width:15px;height:15px;padding:0 4px;border-radius:99px;font-size:9px;font-weight:600;background:var(--panel-2);color:var(--muted)}
.qr-tab--active .qr-tab__badge{background:var(--accent);color:#fff}
.qr-list{display:flex;flex:1;flex-direction:column;gap:8px;min-height:0;overflow-y:auto;padding-right:4px}
.qr-item{display:grid;grid-template-columns:minmax(0,1fr) auto;gap:10px;align-items:start;padding:10px 12px;border:1px solid var(--border);border-radius:10px;background:var(--bg-2)}
.qr-item__main{min-width:0}
.qr-item__pick{display:block;width:100%;padding:0;border:none;background:none;text-align:left;cursor:pointer}
.qr-item__pick:hover .qr-item__name{color:var(--accent)}
.qr-item__name{font-size:13px;font-weight:600;color:var(--fg);margin-bottom:4px;transition:color .15s ease}
.qr-item__body{font-size:12px;line-height:1.5;color:var(--fg-2);white-space:pre-wrap;word-break:break-word}
.qr-item__ops{display:flex;gap:6px;flex-shrink:0}
.qr-op{padding:5px 9px;border-radius:7px;border:1px solid var(--border);background:var(--panel);color:var(--fg-2);font-size:12px;cursor:pointer}
.qr-op:hover{color:var(--fg);border-color:var(--border-strong)}
.qr-op--danger{color:var(--danger)}
.qr-op--danger:hover{border-color:var(--danger);background:var(--danger-soft)}
.qr-empty{display:flex;align-items:center;justify-content:center;min-height:120px;padding:28px 12px;border:1px dashed var(--border);border-radius:10px;text-align:center;color:var(--muted);font-size:13px}
.qr-editor{display:none;padding:14px;border:1px solid var(--border);border-radius:10px;background:var(--bg-2)}
.qr-editor--open{display:block}
.qr-editor__title{font-size:13px;font-weight:600;color:var(--fg);margin-bottom:12px}
.qr-form-row{display:flex;flex-direction:column;gap:5px;margin-bottom:12px}
.qr-form-row label{font-size:12px;font-weight:500;color:var(--fg-2)}
.qr-form-row input,.qr-form-row textarea,.qr-form-row select{width:100%;padding:8px 10px;border-radius:8px;border:1px solid var(--border);background:var(--panel);color:var(--fg);font-size:13px}
.qr-form-row textarea{min-height:92px;resize:vertical}
.qr-form-row input:focus,.qr-form-row textarea:focus,.qr-form-row select:focus{outline:none;border-color:var(--accent)}
.qr-error{display:none;color:var(--danger);font-size:12px;margin-top:-4px;margin-bottom:12px}
.qr-editor__actions,.qr-modal__actions{display:flex;justify-content:flex-end;gap:8px;flex-wrap:wrap}
.qr-btn{padding:7px 14px;border-radius:8px;font-size:13px;cursor:pointer;border:1px solid var(--border);background:var(--panel);color:var(--fg-2)}
.qr-btn:hover{color:var(--fg);border-color:var(--border-strong)}
.qr-btn--primary{background:var(--accent);border-color:var(--accent);color:#fff}
.qr-btn--primary:hover{background:var(--accent-strong);border-color:var(--accent-strong)}
.qr-btn--primary:disabled{background:var(--panel-2);border-color:var(--border);color:var(--muted);cursor:default}
.qr-btn--ghost{background:transparent}
@media(max-width:768px){#qr-modal .modal{width:min(94vw,94vw);max-width:94vw;max-height:90vh}.qr-toolbar{flex-direction:column;align-items:stretch}.qr-toolbar__actions{justify-content:flex-start}.qr-toggle{justify-content:flex-start}.qr-item{grid-template-columns:1fr}.qr-item__ops{justify-content:flex-end;flex-wrap:wrap}}
`;
document.head.appendChild(style);

const composer = document.querySelector('.composer');
const actionHost = document.getElementById('footer-actions');
if (!composer || !composer.parentNode) {
  return;
}

let qrReplies = [];
let categories = [];
let editingIndex = -1;
let saving = false;
let activeCategory = ''; // '' = show all

function escapeHTML(value) {
  return String(value)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function cloneReplies(replies) {
  return (replies || []).map((reply) => ({
    name: reply.name || '',
    body: reply.body || '',
    category: reply.category || ''
  }));
}

function autoSendEnabled() {
  return window.localStorage.getItem(AUTO_SEND_KEY) === '1';
}

function setAutoSendEnabled(on) {
  window.localStorage.setItem(AUTO_SEND_KEY, on ? '1' : '0');
}

function insertReply(body) {
  const text = body || '';
  if (input.value && text) {
    input.value = input.value + '\n' + text;
  } else if (text) {
    input.value = text;
  }
  input.style.height = 'auto';
  input.style.height = Math.min(input.scrollHeight, 140) + 'px';
  input.focus();
  if (autoSendEnabled()) {
    void send();
  }
}

function closeModal() {
  const modal = document.getElementById('qr-modal');
  if (modal) {
    modal.style.display = 'none';
  }
  hideEditor();
}

function setError(message) {
  const error = document.getElementById('qr-error');
  if (!error) {
    return;
  }
  error.textContent = message;
  error.style.display = message ? 'block' : 'none';
}

function hideEditor() {
  editingIndex = -1;
  const editor = document.getElementById('qr-editor');
  if (editor) {
    editor.classList.remove('qr-editor--open');
  }
  const title = document.getElementById('qr-editor-title');
  const name = document.getElementById('qr-name');
  const body = document.getElementById('qr-body');
  if (title) title.textContent = t('新增快捷回复', 'Add Quick Reply');
  if (name) name.value = '';
  if (body) body.value = '';
  setError('');
}

function showEditor(index) {
  editingIndex = typeof index === 'number' ? index : -1;
  const editor = document.getElementById('qr-editor');
  const title = document.getElementById('qr-editor-title');
  const name = document.getElementById('qr-name');
  const body = document.getElementById('qr-body');
  const catSelect = document.getElementById('qr-category');
  if (!editor || !title || !name || !body || !catSelect) {
    return;
  }

  // Build category dropdown options from loaded categories
  catSelect.innerHTML = '';
  const noneOpt = document.createElement('option');
  noneOpt.value = '';
  noneOpt.textContent = t('无分类', 'No category');
  catSelect.appendChild(noneOpt);
  categories.forEach((cat) => {
    const opt = document.createElement('option');
    opt.value = cat.id;
    opt.textContent = lang === 'zh' ? cat.name_zh : cat.name_en;
    catSelect.appendChild(opt);
  });

  if (editingIndex >= 0) {
    const reply = qrReplies[editingIndex];
    title.textContent = t('编辑快捷回复', 'Edit Quick Reply');
    name.value = reply ? reply.name : '';
    body.value = reply ? reply.body : '';
    catSelect.value = reply ? (reply.category || '') : '';
  } else {
    title.textContent = t('新增快捷回复', 'Add Quick Reply');
    name.value = '';
    body.value = '';
    // Default to the currently active category when adding a new reply
    catSelect.value = activeCategory || '';
  }
  setError('');
  editor.classList.add('qr-editor--open');
  name.focus();
}

function renderCategoryTabs() {
  const tabs = document.getElementById('qr-tabs');
  if (!tabs) return;
  tabs.innerHTML = '';

  // "All" tab — always first
  const allTab = document.createElement('button');
  allTab.className = 'qr-tab' + (activeCategory === '' ? ' qr-tab--active' : '');
  allTab.type = 'button';
  allTab.dataset.category = '';
  allTab.innerHTML = t('全部', 'All') + ' <span class="qr-tab__badge">' + qrReplies.length + '</span>';
  tabs.appendChild(allTab);

  // One tab per category
  categories.forEach((cat) => {
    const count = qrReplies.filter((r) => r.category === cat.id).length;
    const tab = document.createElement('button');
    tab.className = 'qr-tab' + (activeCategory === cat.id ? ' qr-tab--active' : '');
    tab.type = 'button';
    tab.dataset.category = cat.id;
    tab.innerHTML = escapeHTML(lang === 'zh' ? cat.name_zh : cat.name_en) + ' <span class="qr-tab__badge">' + count + '</span>';
    tabs.appendChild(tab);
  });
}

function renderList() {
  const list = document.getElementById('qr-list');
  if (!list) {
    return;
  }
  list.innerHTML = '';

  // Map with original index so edit/delete still refer to the correct item
  const indexed = qrReplies.map((reply, idx) => ({ reply, idx }));
  let filtered = indexed;
  if (activeCategory) {
    filtered = filtered.filter((item) => item.reply.category === activeCategory);
  }

  if (!filtered.length) {
    const empty = document.createElement('div');
    empty.className = 'qr-empty';
    empty.textContent = activeCategory
      ? t('此分类暂无快捷回复', 'No quick replies in this category.')
      : t('暂无快捷回复，点击\u201c新增\u201d开始创建', 'No quick replies yet. Click "Add" to create one.');
    list.appendChild(empty);
    return;
  }

  filtered.forEach(({ reply, idx }) => {
    const row = document.createElement('div');
    row.className = 'qr-item';
    row.innerHTML =
      '<div class="qr-item__main">' +
        '<button type="button" class="qr-item__pick" data-action="pick" data-index="' + idx + '">' +
          '<div class="qr-item__name">' + escapeHTML(reply.name) + '</div>' +
          '<div class="qr-item__body">' + escapeHTML(reply.body) + '</div>' +
        '</button>' +
      '</div>' +
      '<div class="qr-item__ops">' +
        '<button type="button" class="qr-op" data-action="edit" data-index="' + idx + '">' + t('编辑', 'Edit') + '</button>' +
        '<button type="button" class="qr-op qr-op--danger" data-action="delete" data-index="' + idx + '">' + t('删除', 'Delete') + '</button>' +
      '</div>';
    list.appendChild(row);
  });
}

function persistReplies(nextReplies, onSuccess) {
  if (saving) {
    return;
  }
  saving = true;
  const saveButton = document.getElementById('qr-save');
  if (saveButton) {
    saveButton.disabled = true;
  }
  setError('');
  fetch('/quick-replies', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(nextReplies)
  }).then((response) => {
    if (!response.ok) {
      throw new Error('save failed');
    }
    qrReplies = cloneReplies(nextReplies);
    renderCategoryTabs();
    renderList();
    if (typeof onSuccess === 'function') {
      onSuccess();
    }
  }).catch((error) => {
    setError(t('保存失败：' + error.message, 'Save failed: ' + error.message));
  }).finally(() => {
    saving = false;
    if (saveButton) {
      saveButton.disabled = false;
    }
  });
}

function saveReply() {
  const nameField = document.getElementById('qr-name');
  const bodyField = document.getElementById('qr-body');
  const catField = document.getElementById('qr-category');
  if (!nameField || !bodyField || !catField) {
    return;
  }
  const name = nameField.value.trim();
  const body = bodyField.value.trim();
  const category = catField.value || '';
  if (!name || !body) {
    setError(t('名称和内容都不能为空', 'Name and message are required'));
    return;
  }
  const nextReplies = cloneReplies(qrReplies);
  const reply = { name, body, category };
  if (editingIndex >= 0) {
    nextReplies[editingIndex] = reply;
  } else {
    nextReplies.push(reply);
  }
  persistReplies(nextReplies, hideEditor);
}

function deleteReply(index) {
  const reply = qrReplies[index];
  if (!reply) {
    return;
  }
  if (!window.confirm(t('确认删除此快捷回复？', 'Delete this quick reply?'))) {
    return;
  }
  const nextReplies = cloneReplies(qrReplies);
  nextReplies.splice(index, 1);
  persistReplies(nextReplies, () => {
    if (editingIndex === index) {
      hideEditor();
    } else if (editingIndex > index) {
      editingIndex -= 1;
    }
  });
}

function handleListClick(event) {
  const target = event.target.closest('[data-action]');
  if (!target) {
    return;
  }
  const index = Number(target.dataset.index);
  const action = target.dataset.action;
  if (action === 'pick') {
    const reply = qrReplies[index];
    if (!reply) {
      return;
    }
    insertReply(reply.body);
    closeModal();
    return;
  }
  if (action === 'edit') {
    showEditor(index);
    return;
  }
  if (action === 'delete') {
    deleteReply(index);
  }
}

function ensureModal() {
  const existing = document.getElementById('qr-modal');
  if (existing) {
    return existing;
  }

  const overlay = document.createElement('div');
  overlay.className = 'modal-overlay';
  overlay.id = 'qr-modal';

  const modal = document.createElement('div');
  modal.className = 'modal';

  const head = document.createElement('div');
  head.className = 'modal__head';
  head.innerHTML =
    '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 7h16"/><path d="M4 12h16"/><path d="M4 17h10"/></svg> ' +
    t('快捷回复', 'Quick Replies') +
    '<span class="modal__close" id="qr-modal-close">&times;</span>';

  const body = document.createElement('div');
  body.className = 'modal__body';
  body.id = 'qr-modal-body';
  body.innerHTML =
    '<div class="qr-shell">' +
      '<div class="qr-toolbar">' +
        '<div class="qr-toolbar__actions">' +
          '<button type="button" class="qr-btn qr-btn--primary" id="qr-add">' + t('新增', 'Add') + '</button>' +
        '</div>' +
        '<label class="qr-toggle"><input type="checkbox" id="qr-auto-send" />' + t('点击后自动发送', 'Auto-send after insert') + '</label>' +
      '</div>' +
      '<div class="qr-tabs" id="qr-tabs"></div>' +
      '<div class="qr-list" id="qr-list"></div>' +
      '<div class="qr-editor" id="qr-editor">' +
        '<div class="qr-editor__title" id="qr-editor-title"></div>' +
        '<div class="qr-form-row"><label>' + t('名称', 'Name') + '</label><input id="qr-name" /></div>' +
        '<div class="qr-form-row"><label>' + t('分类', 'Category') + '</label><select id="qr-category"></select></div>' +
        '<div class="qr-form-row"><label>' + t('消息内容', 'Message') + '</label><textarea id="qr-body"></textarea></div>' +
        '<div class="qr-error" id="qr-error"></div>' +
        '<div class="qr-editor__actions">' +
          '<button type="button" class="qr-btn qr-btn--ghost" id="qr-cancel-edit">' + t('取消', 'Cancel') + '</button>' +
          '<button type="button" class="qr-btn qr-btn--primary" id="qr-save">' + t('保存', 'Save') + '</button>' +
        '</div>' +
      '</div>' +
      '<div class="qr-modal__actions">' +
        '<button type="button" class="qr-btn" id="qr-close">' + t('关闭', 'Close') + '</button>' +
      '</div>' +
    '</div>';

  modal.appendChild(head);
  modal.appendChild(body);
  overlay.appendChild(modal);
  document.body.appendChild(overlay);

  document.getElementById('qr-modal-close').onclick = closeModal;
  document.getElementById('qr-close').onclick = closeModal;
  document.getElementById('qr-add').onclick = () => showEditor(-1);
  document.getElementById('qr-cancel-edit').onclick = hideEditor;
  document.getElementById('qr-save').onclick = saveReply;
  document.getElementById('qr-list').onclick = handleListClick;
  document.getElementById('qr-auto-send').onchange = (event) => {
    setAutoSendEnabled(Boolean(event.target.checked));
  };
  document.getElementById('qr-tabs').onclick = (event) => {
    const tab = event.target.closest('.qr-tab');
    if (!tab) return;
    activeCategory = tab.dataset.category || '';
    renderCategoryTabs();
    renderList();
  };
  overlay.onclick = (event) => {
    if (event.target === overlay) {
      closeModal();
    }
  };

  return overlay;
}

function openComposerPicker() {
  const modal = ensureModal();
  const toggle = document.getElementById('qr-auto-send');
  if (toggle) {
    toggle.checked = autoSendEnabled();
  }
  renderCategoryTabs();
  renderList();
  hideEditor();
  modal.style.display = 'flex';
}

function loadReplies() {
  return fetch('/quick-replies')
    .then((response) => {
      if (!response.ok) {
        throw new Error('load failed');
      }
      return response.json();
    })
    .then((replies) => {
      qrReplies = cloneReplies(Array.isArray(replies) ? replies : []);
      renderCategoryTabs();
      renderList();
    })
    .catch(() => {
      qrReplies = [];
      renderCategoryTabs();
      renderList();
    });
}

function loadCategories() {
  return fetch('/quick-reply-categories')
    .then((response) => {
      if (!response.ok) throw new Error('load categories failed');
      return response.json();
    })
    .then((data) => {
      categories = Array.isArray(data) ? data : [];
      renderCategoryTabs();
    })
    .catch(() => {
      categories = [];
      renderCategoryTabs();
    });
}

const launch = document.createElement('button');
launch.type = 'button';
launch.className = 'footer-action qr-launch';
launch.style.order = '20';
launch.innerHTML =
  '<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 7h16"/><path d="M4 12h16"/><path d="M4 17h10"/></svg>' +
  '<span>' + t('快捷回复', 'Quick Replies') + '</span>';
launch.onclick = () => {
  openComposerPicker();
};
if (actionHost) {
  actionHost.appendChild(launch);
} else {
  composer.appendChild(launch);
}

Promise.all([loadReplies(), loadCategories()]).catch(function() {
  // both already call render functions on error
});
})();
