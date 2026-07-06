(function() {
'use strict';

const lang = (document.documentElement.lang || 'en').startsWith('zh') ? 'zh' : 'en';
const t = (zh, en) => (lang === 'zh' ? zh : en);

const style = document.createElement('style');
style.textContent =
`.qr-bar{display:flex;gap:5px;padding:3px 0 0;flex-wrap:wrap;align-items:center}
.qr-bar--empty{justify-content:flex-end}
.qr-btn{font-size:12px;padding:3px 9px;border-radius:5px;background:var(--bg-2);border:1px solid var(--border);color:var(--fg-2);transition:all .15s;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;max-width:180px;cursor:pointer}
.qr-btn:hover{background:var(--panel);color:var(--fg);border-color:var(--border-strong)}
.qr-btn--send{background:var(--accent-soft);border-color:var(--accent);color:var(--accent)}
.qr-btn--send:hover{background:var(--accent);color:#fff}
.qr-btn--mgmt{font-size:14px;padding:3px 7px;margin-left:auto;opacity:.5}
.qr-btn--mgmt:hover{opacity:1}
.qr-form-row{display:flex;flex-direction:column;gap:4px;margin-bottom:12px}
.qr-form-row label{font-size:12px;font-weight:500;color:var(--fg-2)}
.qr-form-row input,.qr-form-row textarea{padding:7px 10px;border-radius:var(--radius);border:1px solid var(--border);background:var(--bg-2);color:var(--fg);font-size:13px;width:100%}
.qr-form-row textarea{resize:vertical;min-height:60px}
.qr-form-row input:focus,.qr-form-row textarea:focus{border-color:var(--accent);outline:none}
.qr-form-row--inline{flex-direction:row;align-items:center;gap:8px}
.qr-form-row input[type=checkbox]{width:auto}
.qr-item{display:flex;align-items:center;gap:8px;padding:8px 10px;border-radius:var(--radius);background:var(--bg-2);border:1px solid var(--border);margin-bottom:6px}
.qr-item__label{flex:1;font-size:13px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.qr-item__badge{font-size:10px;padding:1px 6px;border-radius:4px;background:var(--accent-soft);color:var(--accent)}
.qr-item__del{padding:3px 8px;border-radius:4px;font-size:12px;color:var(--danger);cursor:pointer;transition:background .15s;border:none;background:none}
.qr-item__del:hover{background:var(--danger-soft)}
.qr-empty{padding:20px 0;text-align:center;color:var(--muted);font-size:13px}
.qr-actions{display:flex;gap:8px;justify-content:flex-end;margin-top:16px;padding-top:12px;border-top:1px solid var(--border)}
.qr-actions button{padding:7px 16px;border-radius:var(--radius);font-size:13px;font-weight:500;cursor:pointer;transition:all .15s}
.qr-btn-primary{background:var(--accent);color:#fff;border:none}
.qr-btn-primary:hover{background:var(--accent-strong)}
.qr-btn-primary:disabled{background:var(--panel-2);color:var(--muted);cursor:default}
.qr-btn-secondary{background:var(--bg-2);color:var(--fg-2);border:1px solid var(--border)}
.qr-btn-secondary:hover{background:var(--panel);color:var(--fg)}
.qr-add-section{margin-top:16px;padding-top:12px;border-top:1px dashed var(--border)}
.qr-add-section h4{font-size:13px;font-weight:600;margin-bottom:10px;color:var(--fg-2)}
.qr-error{color:var(--danger);font-size:12px;margin-top:4px;display:none}`;
document.head.appendChild(style);

let qrReplies = [];
let modalReplies = null;

const composer = document.querySelector('.composer');
if (!composer || !composer.parentNode) {
  return;
}

const bar = document.createElement('div');
bar.id = 'qr-bar';
bar.className = 'qr-bar qr-bar--empty';
composer.parentNode.insertBefore(bar, composer);

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
    autoSend: Boolean(reply.autoSend),
    icon: reply.icon || ''
  }));
}

function renderBar() {
  bar.innerHTML = '';
  bar.classList.toggle('qr-bar--empty', qrReplies.length === 0);

  const frag = document.createDocumentFragment();
  qrReplies.forEach((reply) => {
    const btn = document.createElement('button');
    btn.className = 'qr-btn' + (reply.autoSend ? ' qr-btn--send' : '');
    btn.textContent = (reply.icon || '▸') + ' ' + reply.name;
    btn.title = reply.body;
    btn.addEventListener('click', (event) => {
      event.stopPropagation();
      input.value = reply.body;
      input.style.height = 'auto';
      input.style.height = Math.min(input.scrollHeight, 140) + 'px';
      if (reply.autoSend) {
        void send();
        return;
      }
      input.focus();
    });
    frag.appendChild(btn);
  });

  const manage = document.createElement('button');
  manage.className = 'qr-btn qr-btn--mgmt';
  manage.textContent = '⚙';
  manage.title = t('管理快捷回复', 'Manage quick replies');
  manage.addEventListener('click', openModal);
  frag.appendChild(manage);

  bar.appendChild(frag);
  bar.style.display = 'flex';
}

function loadReplies() {
  fetch('/quick-replies')
    .then((response) => {
      if (!response.ok) {
        throw new Error('load failed');
      }
      return response.json();
    })
    .then((replies) => {
      qrReplies = Array.isArray(replies) ? replies : [];
      renderBar();
    })
    .catch(() => {
      qrReplies = [];
      renderBar();
    });
}

function openModal() {
  modalReplies = cloneReplies(qrReplies);
  const existing = document.getElementById('qr-modal');
  if (existing) {
    const modalBody = document.getElementById('qr-modal-body');
    if (modalBody) {
      buildModalBody(modalBody);
    }
    existing.style.display = 'flex';
    return;
  }

  const overlay = document.createElement('div');
  overlay.className = 'modal-overlay';
  overlay.id = 'qr-modal';

  const modal = document.createElement('div');
  modal.className = 'modal';

  const head = document.createElement('div');
  head.className = 'modal__head';
  head.innerHTML = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20V10"/><path d="M18 20V4"/><path d="M6 20v-6"/></svg> ' +
    t('快捷回复', 'Quick Replies') +
    '<span class="modal__close" id="qr-modal-close">&times;</span>';

  const body = document.createElement('div');
  body.className = 'modal__body';
  body.id = 'qr-modal-body';
  buildModalBody(body);

  modal.appendChild(head);
  modal.appendChild(body);
  overlay.appendChild(modal);
  document.body.appendChild(overlay);

  setTimeout(() => {
    overlay.style.display = 'flex';
  }, 10);

  $('#qr-modal-close').onclick = () => {
    overlay.style.display = 'none';
  };
  overlay.onclick = (event) => {
    if (event.target === overlay) {
      overlay.style.display = 'none';
    }
  };
}

function setError(message) {
  const error = $('#qr-error');
  if (!error) {
    return;
  }
  error.textContent = message;
  error.style.display = message ? 'block' : 'none';
}

function draftReply() {
  return {
    name: $('#qr-new-name').value.trim(),
    body: $('#qr-new-body').value.trim(),
    autoSend: $('#qr-new-auto').checked,
    icon: $('#qr-new-icon').value.trim()
  };
}

function persistReplies(nextReplies, options) {
  const opts = options || {};
  const saveButton = $('#qr-save');
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
    qrReplies = nextReplies;
    renderBar();
    if (typeof opts.onSuccess === 'function') {
      opts.onSuccess();
    }
  }).catch((error) => {
    setError(t('保存失败：' + error.message, 'Save failed: ' + error.message));
  }).finally(() => {
    if (saveButton) {
      saveButton.disabled = false;
    }
  });
}

function saveChanges() {
  const draft = draftReply();
  const hasName = draft.name !== '';
  const hasBody = draft.body !== '';
  if (hasName !== hasBody) {
    setError(t('请同时填写名称和消息内容，或留空仅保存现有修改', 'Fill in both name and message, or leave both empty to save existing changes only'));
    return;
  }

  const nextReplies = cloneReplies(modalReplies);
  if (hasName && hasBody) {
    nextReplies.push(draft);
  }

  persistReplies(nextReplies, {
    onSuccess() {
      modalReplies = null;
      const modal = document.getElementById('qr-modal');
      if (modal) {
        modal.style.display = 'none';
      }
    }
  });
}

function buildModalBody(body) {
  body.innerHTML = '';
  const draftReplies = cloneReplies(modalReplies);

  const list = document.createElement('div');
  list.id = 'qr-list';
  if (!draftReplies.length) {
    const empty = document.createElement('div');
    empty.className = 'qr-empty';
    empty.textContent = t('暂无快捷回复，请在下方添加', 'No quick replies yet - add one below');
    list.appendChild(empty);
  } else {
    draftReplies.forEach((reply, index) => {
      const item = document.createElement('div');
      item.className = 'qr-item';
      item.innerHTML = '<span class="qr-item__label">' + escapeHTML(reply.icon || '') + ' ' + escapeHTML(reply.name) + '</span>' +
        (reply.autoSend ? '<span class="qr-item__badge">' + t('自动', 'Auto') + '</span>' : '') +
        '<button class="qr-item__del" data-idx="' + index + '">' + t('删除', 'Delete') + '</button>';
      item.querySelector('.qr-item__del').addEventListener('click', () => {
        if (!confirm(t('确认删除此快捷回复？', 'Delete this quick reply?'))) {
          return;
        }
        modalReplies.splice(index, 1);
        buildModalBody(body);
      });
      list.appendChild(item);
    });
  }
  body.appendChild(list);

  const addSection = document.createElement('div');
  addSection.className = 'qr-add-section';
  addSection.innerHTML = '<h4>' + t('添加快捷回复', 'Add Quick Reply') + '</h4>' +
    '<div class="qr-form-row"><label>' + t('名称', 'Name') + '</label><input id="qr-new-name" /></div>' +
    '<div class="qr-form-row"><label>' + t('消息内容', 'Message') + '</label><textarea id="qr-new-body" rows="3"></textarea></div>' +
    '<div class="qr-form-row qr-form-row--inline"><label>' + t('图标', 'Icon') + '</label><input id="qr-new-icon" style="width:60px" /></div>' +
    '<div class="qr-form-row qr-form-row--inline"><input type="checkbox" id="qr-new-auto" checked /><label>' + t('自动发送', 'Auto-send') + '</label></div>' +
    '<div class="qr-error" id="qr-error"></div>';
  body.appendChild(addSection);

  const actions = document.createElement('div');
  actions.className = 'qr-actions';
  actions.innerHTML = '<button class="qr-btn-secondary" id="qr-cancel">' + t('取消', 'Cancel') + '</button>' +
    '<button class="qr-btn-primary" id="qr-save">' + t('保存', 'Save') + '</button>';
  body.appendChild(actions);

  $('#qr-cancel').onclick = () => {
    modalReplies = null;
    const modal = document.getElementById('qr-modal');
    if (modal) {
      modal.style.display = 'none';
    }
  };
  $('#qr-save').onclick = saveChanges;
}

renderBar();
loadReplies();
})();
