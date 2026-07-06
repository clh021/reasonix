(function() {
'use strict';

const lang = (document.documentElement.lang || 'en').startsWith('zh') ? 'zh' : 'en';
const t = (zh, en) => (lang === 'zh' ? zh : en);
const FEATURE_STATUSES = ['wishlist', 'planned', 'in_progress', 'custom_active', 'upstream_covered', 'dropped'];
const FEATURE_PRIORITIES = ['high', 'medium', 'low'];

const STATUS_LABELS = {
  en: {
    wishlist: 'Wishlist',
    planned: 'Planned',
    in_progress: 'In Progress',
    custom_active: 'Custom Active',
    upstream_covered: 'Upstream Covered',
    dropped: 'Dropped'
  },
  zh: {
    wishlist: '希望以后支持',
    planned: '已计划',
    in_progress: '进行中',
    custom_active: '当前自维护',
    upstream_covered: '已被上游覆盖',
    dropped: '已放弃'
  }
};

const PRIORITY_LABELS = {
  en: {high: 'High', medium: 'Medium', low: 'Low'},
  zh: {high: '高', medium: '中', low: '低'}
};

const style = document.createElement('style');
style.textContent =
`.ft-shell{display:flex;flex:1;flex-direction:column;gap:14px;width:100%;min-width:0;min-height:0}
#ft-modal .modal{display:flex;flex-direction:column;width:min(960px,92vw);max-width:92vw;max-height:min(88vh,960px);overflow:hidden}
#ft-modal .modal__body{display:flex;flex:1;min-height:0;overflow:hidden}
.ft-headline{display:flex;justify-content:space-between;gap:12px;align-items:flex-start;flex-wrap:wrap}
.ft-headline__meta{display:flex;flex:1;flex-direction:column;gap:4px;min-width:0}
.ft-headline__title{font-size:14px;font-weight:600;color:var(--fg)}
.ft-headline__path{font-family:var(--mono);font-size:11px;color:var(--muted);word-break:break-all}
.ft-actions{display:flex;align-items:center;gap:8px;flex-wrap:wrap}
.ft-btn{padding:7px 14px;border-radius:8px;border:1px solid var(--border);background:var(--panel);color:var(--fg-2);font-size:13px;cursor:pointer}
.ft-btn:hover{color:var(--fg);border-color:var(--border-strong)}
.ft-btn--primary{background:var(--accent);border-color:var(--accent);color:#fff}
.ft-btn--primary:hover{background:var(--accent-strong);border-color:var(--accent-strong)}
.ft-btn--primary:disabled{background:var(--panel-2);border-color:var(--border);color:var(--muted);cursor:default}
.ft-list{display:flex;flex:1;flex-direction:column;gap:14px;min-height:0;overflow-y:auto;padding-right:4px}
.ft-section{display:flex;flex-direction:column;gap:8px}
.ft-section__title{font-size:12px;font-weight:700;color:var(--fg-2);text-transform:uppercase;letter-spacing:.04em}
.ft-empty{padding:20px 12px;border:1px dashed var(--border);border-radius:10px;text-align:center;color:var(--muted);font-size:13px}
.ft-card{display:flex;justify-content:space-between;gap:12px;padding:12px;border:1px solid var(--border);border-radius:10px;background:var(--bg-2)}
.ft-card__main{min-width:0;display:flex;flex-direction:column;gap:7px}
.ft-card__title{font-size:14px;font-weight:600;color:var(--fg)}
.ft-card__summary{font-size:12px;line-height:1.55;color:var(--fg-2);white-space:pre-wrap;word-break:break-word}
.ft-card__notes{font-size:12px;line-height:1.55;color:var(--muted);white-space:pre-wrap;word-break:break-word}
.ft-badges{display:flex;gap:6px;flex-wrap:wrap}
.ft-badge{display:inline-flex;align-items:center;padding:2px 7px;border-radius:999px;font-size:10.5px;border:1px solid var(--border);background:var(--panel);color:var(--fg-2)}
.ft-badge--priority-high{background:var(--danger-soft);border-color:transparent;color:var(--danger)}
.ft-badge--priority-medium{background:var(--warning-soft);border-color:transparent;color:var(--warning)}
.ft-badge--priority-low{background:var(--success-soft);border-color:transparent;color:var(--success)}
.ft-card__ops{display:flex;gap:6px;flex-shrink:0;align-items:flex-start}
.ft-op{padding:6px 10px;border-radius:8px;border:1px solid var(--border);background:var(--panel);font-size:12px;color:var(--fg-2);cursor:pointer}
.ft-op:hover{color:var(--fg);border-color:var(--border-strong)}
.ft-op--danger{color:var(--danger)}
.ft-op--danger:hover{background:var(--danger-soft);border-color:var(--danger)}
.ft-editor{display:none;padding:14px;border:1px solid var(--border);border-radius:10px;background:var(--bg-2)}
.ft-editor--open{display:block}
.ft-editor__title{font-size:13px;font-weight:600;color:var(--fg);margin-bottom:12px}
.ft-form-row{display:flex;flex-direction:column;gap:5px;margin-bottom:12px}
.ft-form-row label{font-size:12px;font-weight:500;color:var(--fg-2)}
.ft-form-row input,.ft-form-row textarea,.ft-form-row select{width:100%;padding:8px 10px;border-radius:8px;border:1px solid var(--border);background:var(--panel);color:var(--fg);font-size:13px}
.ft-form-row textarea{min-height:80px;resize:vertical}
.ft-form-row input:focus,.ft-form-row textarea:focus,.ft-form-row select:focus{outline:none;border-color:var(--accent)}
.ft-grid{display:grid;grid-template-columns:1fr 180px 140px;gap:10px}
.ft-error{display:none;color:var(--danger);font-size:12px;margin-top:-4px;margin-bottom:12px}
.ft-editor__actions{display:flex;justify-content:flex-end;gap:8px;flex-wrap:wrap}
@media(max-width:768px){#ft-modal .modal{width:min(94vw,94vw);max-width:94vw;max-height:90vh}.ft-headline{flex-direction:column;align-items:stretch}.ft-actions{justify-content:flex-start}.ft-grid{grid-template-columns:1fr}.ft-card{flex-direction:column}.ft-card__ops{justify-content:flex-end;flex-wrap:wrap}}
`;
document.head.appendChild(style);

const composer = document.querySelector('.composer');
const actionHost = document.getElementById('footer-actions');
if (!composer || !composer.parentNode) {
  return;
}

let trackerProject = null;
let trackerFeatures = [];
let editingFeatureId = '';
let loading = false;
let loadError = '';
let loadingFeatures = false;

function statusLabel(status) {
  return STATUS_LABELS[lang]?.[status] || STATUS_LABELS.en[status] || status;
}

function priorityLabel(priority) {
  return PRIORITY_LABELS[lang]?.[priority] || PRIORITY_LABELS.en[priority] || priority;
}

function escapeHTML(value) {
  return String(value)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function cloneFeatures(features) {
  return (features || []).map((feature) => ({
    id: feature.id || '',
    title: feature.title || '',
    status: feature.status || 'wishlist',
    priority: feature.priority || 'medium',
    summary: feature.summary || '',
    notes: feature.notes || ''
  }));
}

function setFeatureTrackerError(message) {
  const error = document.getElementById('ft-error');
  if (!error) return;
  error.textContent = message;
  error.style.display = message ? 'block' : 'none';
}

function hideFeatureEditor() {
  editingFeatureId = '';
  const editor = document.getElementById('ft-editor');
  if (!editor) return;
  editor.classList.remove('ft-editor--open');
  document.getElementById('ft-title').value = '';
  document.getElementById('ft-status').value = 'wishlist';
  document.getElementById('ft-priority').value = 'medium';
  document.getElementById('ft-summary').value = '';
  document.getElementById('ft-notes').value = '';
  document.getElementById('ft-editor-title').textContent = t('新增特性', 'Add Feature');
  setFeatureTrackerError('');
}

function showFeatureEditor(feature) {
  const editor = document.getElementById('ft-editor');
  if (!editor) return;
  editingFeatureId = feature?.id || '';
  document.getElementById('ft-title').value = feature?.title || '';
  document.getElementById('ft-status').value = feature?.status || 'wishlist';
  document.getElementById('ft-priority').value = feature?.priority || 'medium';
  document.getElementById('ft-summary').value = feature?.summary || '';
  document.getElementById('ft-notes').value = feature?.notes || '';
  document.getElementById('ft-editor-title').textContent = editingFeatureId ? t('编辑特性', 'Edit Feature') : t('新增特性', 'Add Feature');
  setFeatureTrackerError('');
  editor.classList.add('ft-editor--open');
  document.getElementById('ft-title').focus();
}

function renderFeatureSections() {
  const list = document.getElementById('ft-list');
  if (!list) return;
  list.innerHTML = '';
  if (loadingFeatures) {
    const empty = document.createElement('div');
    empty.className = 'ft-empty';
    empty.textContent = t('加载中...', 'Loading...');
    list.appendChild(empty);
    return;
  }
  if (loadError) {
    const empty = document.createElement('div');
    empty.className = 'ft-empty';
    empty.textContent = loadError;
    list.appendChild(empty);
    return;
  }
  if (!trackerFeatures.length) {
    const empty = document.createElement('div');
    empty.className = 'ft-empty';
    empty.textContent = t('当前项目还没有特性条目，点击“新增特性”开始记录。', 'No features yet for this project. Click "Add Feature" to start tracking.');
    list.appendChild(empty);
    return;
  }
  FEATURE_STATUSES.forEach((status) => {
    const items = trackerFeatures.filter((feature) => feature.status === status);
    if (!items.length) return;
    const section = document.createElement('section');
    section.className = 'ft-section';
    section.innerHTML = '<div class="ft-section__title">' + escapeHTML(statusLabel(status)) + '</div>';
    items.forEach((feature) => {
      const card = document.createElement('div');
      card.className = 'ft-card';
      card.innerHTML =
        '<div class="ft-card__main">' +
          '<div class="ft-card__title">' + escapeHTML(feature.title) + '</div>' +
          '<div class="ft-badges">' +
            '<span class="ft-badge">' + escapeHTML(statusLabel(feature.status)) + '</span>' +
            '<span class="ft-badge ft-badge--priority-' + escapeHTML(feature.priority) + '">' + escapeHTML(priorityLabel(feature.priority)) + '</span>' +
          '</div>' +
          (feature.summary ? '<div class="ft-card__summary">' + escapeHTML(feature.summary) + '</div>' : '') +
          (feature.notes ? '<div class="ft-card__notes">' + escapeHTML(feature.notes) + '</div>' : '') +
        '</div>' +
        '<div class="ft-card__ops">' +
          '<button type="button" class="ft-op" data-action="edit" data-id="' + escapeHTML(feature.id) + '">' + t('编辑', 'Edit') + '</button>' +
          '<button type="button" class="ft-op ft-op--danger" data-action="delete" data-id="' + escapeHTML(feature.id) + '">' + t('删除', 'Delete') + '</button>' +
        '</div>';
      section.appendChild(card);
    });
    list.appendChild(section);
  });
}

function renderFeatureTrackerMeta() {
  const title = document.getElementById('ft-project-title');
  const path = document.getElementById('ft-project-path');
  if (!title || !path) return;
  title.textContent = trackerProject?.name || t('当前项目', 'Current Project');
  path.textContent = trackerProject?.root || '';
}

function persistFeatures(nextFeatures, onSuccess) {
  if (loading) return;
  loading = true;
  const saveButton = document.getElementById('ft-save');
  if (saveButton) saveButton.disabled = true;
  setFeatureTrackerError('');
  fetch('/project-features', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({features: nextFeatures})
  }).then((response) => {
    if (!response.ok) throw new Error('save failed');
    return response.json();
  }).then((snapshot) => {
    trackerProject = snapshot.project || trackerProject;
    trackerFeatures = cloneFeatures(snapshot.features);
    renderFeatureTrackerMeta();
    renderFeatureSections();
    if (typeof onSuccess === 'function') onSuccess();
  }).catch((error) => {
    setFeatureTrackerError(t('保存失败：' + error.message, 'Save failed: ' + error.message));
  }).finally(() => {
    loading = false;
    if (saveButton) saveButton.disabled = false;
  });
}

function saveFeature() {
  const title = document.getElementById('ft-title').value.trim();
  const status = document.getElementById('ft-status').value;
  const priority = document.getElementById('ft-priority').value;
  const summary = document.getElementById('ft-summary').value.trim();
  const notes = document.getElementById('ft-notes').value.trim();
  if (!title) {
    setFeatureTrackerError(t('标题不能为空', 'Title is required'));
    return;
  }
  const nextFeatures = cloneFeatures(trackerFeatures);
  const next = {id: editingFeatureId, title, status, priority, summary, notes};
  const index = nextFeatures.findIndex((feature) => feature.id === editingFeatureId);
  if (editingFeatureId && index >= 0) {
    nextFeatures[index] = next;
  } else {
    nextFeatures.push(next);
  }
  persistFeatures(nextFeatures, hideFeatureEditor);
}

function deleteFeature(id) {
  const feature = trackerFeatures.find((item) => item.id === id);
  if (!feature) return;
  if (!window.confirm(t('确认删除这个特性条目？', 'Delete this feature entry?'))) return;
  const nextFeatures = cloneFeatures(trackerFeatures).filter((item) => item.id !== id);
  persistFeatures(nextFeatures, () => {
    if (editingFeatureId === id) hideFeatureEditor();
  });
}

function handleFeatureListClick(event) {
  const btn = event.target.closest('[data-action]');
  if (!btn) return;
  const id = btn.dataset.id || '';
  if (btn.dataset.action === 'edit') {
    const feature = trackerFeatures.find((item) => item.id === id);
    if (feature) showFeatureEditor(feature);
    return;
  }
  if (btn.dataset.action === 'delete') {
    deleteFeature(id);
  }
}

function loadFeatureTracker() {
  loadingFeatures = true;
  loadError = '';
  trackerProject = {name: t('加载中...', 'Loading...'), root: ''};
  trackerFeatures = [];
  renderFeatureTrackerMeta();
  renderFeatureSections();
  return fetch('/project-features')
    .then((response) => {
      if (!response.ok) throw new Error('load failed');
      return response.json();
    })
    .then((snapshot) => {
      loadingFeatures = false;
      loadError = '';
      trackerProject = snapshot.project || null;
      trackerFeatures = cloneFeatures(snapshot.features);
      renderFeatureTrackerMeta();
      renderFeatureSections();
    })
    .catch(() => {
      loadingFeatures = false;
      loadError = t('加载项目特性失败，请稍后重试。', 'Failed to load project features. Please try again.');
      renderFeatureTrackerMeta();
      renderFeatureSections();
    });
}

function ensureFeatureTrackerModal() {
  const existing = document.getElementById('ft-modal');
  if (existing) return existing;

  const overlay = document.createElement('div');
  overlay.className = 'modal-overlay';
  overlay.id = 'ft-modal';
  overlay.style.display = 'none';

  const modal = document.createElement('div');
  modal.className = 'modal';

  const head = document.createElement('div');
  head.className = 'modal__head';
  head.innerHTML =
    '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>' +
    t('项目特性', 'Project Features') +
    '<span class="modal__close" id="ft-close-x">&times;</span>';

  const body = document.createElement('div');
  body.className = 'modal__body';
  body.id = 'ft-modal-body';
  body.innerHTML =
    '<div class="ft-shell">' +
      '<div class="ft-headline">' +
        '<div class="ft-headline__meta">' +
          '<div class="ft-headline__title" id="ft-project-title"></div>' +
          '<div class="ft-headline__path" id="ft-project-path"></div>' +
        '</div>' +
        '<div class="ft-actions">' +
          '<button type="button" class="ft-btn ft-btn--primary" id="ft-add">' + t('新增特性', 'Add Feature') + '</button>' +
          '<button type="button" class="ft-btn" id="ft-refresh">' + t('刷新', 'Refresh') + '</button>' +
        '</div>' +
      '</div>' +
      '<div class="ft-list" id="ft-list"></div>' +
      '<div class="ft-editor" id="ft-editor">' +
        '<div class="ft-editor__title" id="ft-editor-title"></div>' +
        '<div class="ft-grid">' +
          '<div class="ft-form-row"><label>' + t('标题', 'Title') + '</label><input id="ft-title" /></div>' +
          '<div class="ft-form-row"><label>' + t('状态', 'Status') + '</label><select id="ft-status"></select></div>' +
          '<div class="ft-form-row"><label>' + t('优先级', 'Priority') + '</label><select id="ft-priority"></select></div>' +
        '</div>' +
        '<div class="ft-form-row"><label>' + t('简述', 'Summary') + '</label><textarea id="ft-summary"></textarea></div>' +
        '<div class="ft-form-row"><label>' + t('备注', 'Notes') + '</label><textarea id="ft-notes"></textarea></div>' +
        '<div class="ft-error" id="ft-error"></div>' +
        '<div class="ft-editor__actions">' +
          '<button type="button" class="ft-btn" id="ft-cancel">' + t('取消', 'Cancel') + '</button>' +
          '<button type="button" class="ft-btn ft-btn--primary" id="ft-save">' + t('保存', 'Save') + '</button>' +
        '</div>' +
      '</div>' +
      '<div class="ft-editor__actions">' +
        '<button type="button" class="ft-btn" id="ft-close">' + t('关闭', 'Close') + '</button>' +
      '</div>' +
    '</div>';

  modal.appendChild(head);
  modal.appendChild(body);
  overlay.appendChild(modal);
  document.body.appendChild(overlay);

  const statusSelect = document.getElementById('ft-status');
  FEATURE_STATUSES.forEach((status) => {
    const opt = document.createElement('option');
    opt.value = status;
    opt.textContent = statusLabel(status);
    statusSelect.appendChild(opt);
  });
  const prioritySelect = document.getElementById('ft-priority');
  FEATURE_PRIORITIES.forEach((priority) => {
    const opt = document.createElement('option');
    opt.value = priority;
    opt.textContent = priorityLabel(priority);
    prioritySelect.appendChild(opt);
  });

  document.getElementById('ft-close-x').onclick = closeFeatureTracker;
  document.getElementById('ft-close').onclick = closeFeatureTracker;
  document.getElementById('ft-cancel').onclick = hideFeatureEditor;
  document.getElementById('ft-save').onclick = saveFeature;
  document.getElementById('ft-add').onclick = () => showFeatureEditor(null);
  document.getElementById('ft-refresh').onclick = () => void loadFeatureTracker();
  document.getElementById('ft-list').onclick = handleFeatureListClick;
  overlay.onclick = (event) => {
    if (event.target === overlay) closeFeatureTracker();
  };

  return overlay;
}

function closeFeatureTracker() {
  const modal = document.getElementById('ft-modal');
  if (modal) modal.style.display = 'none';
  hideFeatureEditor();
}

function openFeatureTracker() {
  const modal = ensureFeatureTrackerModal();
  modal.style.display = 'flex';
  hideFeatureEditor();
  void loadFeatureTracker();
}

const launch = document.createElement('button');
launch.type = 'button';
launch.className = 'footer-action ft-launch';
launch.style.order = '10';
launch.innerHTML =
  '<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>' +
  '<span>' + t('特性', 'Features') + '</span>';
launch.onclick = openFeatureTracker;
if (actionHost) {
  actionHost.appendChild(launch);
} else {
  const sendButton = document.getElementById('btn-send');
  if (sendButton) {
    composer.insertBefore(launch, sendButton);
  } else {
    composer.appendChild(launch);
  }
}
})();
