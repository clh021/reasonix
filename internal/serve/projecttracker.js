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
`.ft-shell{display:flex;flex:1;flex-direction:column;gap:6px;width:100%;min-width:0;min-height:0}
#ft-modal .modal{display:flex;flex-direction:column;width:min(960px,92vw);max-width:92vw;max-height:min(88vh,960px);overflow:hidden}
#ft-modal .modal__body{display:flex;flex:1;min-height:0;overflow:hidden}
.ft-topbar{display:flex;align-items:center;gap:8px;padding:0 0 4px}
.ft-search{display:flex;flex:1;align-items:center;gap:4px;padding:4px 8px;border:1px solid var(--border);border-radius:6px;background:var(--panel);min-width:0}
.ft-search:focus-within{border-color:var(--accent)}
.ft-search__icon{width:13px;height:13px;color:var(--muted-2);flex-shrink:0}
.ft-search__input{flex:1;border:none;background:none;outline:none;font-size:12px;color:var(--fg);min-width:0}
.ft-search__input::placeholder{color:var(--muted-2)}
.ft-btn{padding:4px 10px;border-radius:6px;border:1px solid var(--border);background:var(--panel);color:var(--fg-2);font-size:12px;cursor:pointer;white-space:nowrap}
.ft-btn:hover{color:var(--fg);border-color:var(--border-strong)}
.ft-btn--primary{background:var(--accent);border-color:var(--accent);color:#fff}
.ft-btn--primary:hover{background:var(--accent-strong);border-color:var(--accent-strong)}
.ft-btn--primary:disabled{background:var(--panel-2);border-color:var(--border);color:var(--muted);cursor:default}
.ft-tabs{display:flex;gap:2px;padding:0;margin-bottom:4px;flex-wrap:nowrap;overflow-x:auto;scrollbar-width:none}
.ft-tabs::-webkit-scrollbar{display:none}
.ft-tab{display:inline-flex;align-items:center;gap:4px;padding:3px 8px;border-radius:5px;font-size:11px;font-weight:500;color:var(--fg-2);cursor:pointer;white-space:nowrap;border:none;background:none;transition:background .15s,color .15s}
.ft-tab:hover{background:var(--panel);color:var(--fg)}
.ft-tab--active{background:var(--accent-soft);color:var(--accent)}
.ft-tab__badge{display:inline-flex;align-items:center;justify-content:center;min-width:15px;height:15px;padding:0 4px;border-radius:99px;font-size:9px;font-weight:600;background:var(--panel-2);color:var(--muted)}
.ft-tab--active .ft-tab__badge{background:var(--accent);color:#fff}
.ft-list{display:flex;flex:1;flex-direction:column;gap:1px;min-height:0;overflow-y:auto;padding-right:2px}
.ft-empty{padding:16px 10px;border:1px dashed var(--border);border-radius:8px;text-align:center;color:var(--muted);font-size:12px}
.ft-row{display:flex;align-items:flex-start;gap:6px;padding:4px 6px;border-radius:5px;cursor:pointer;transition:background .12s ease}
.ft-row:hover{background:var(--panel)}
.ft-row__main{flex:1;min-width:0;display:flex;flex-direction:column;gap:1px}
.ft-row__title{display:inline-flex;align-items:center;gap:4px;font-size:12.5px;font-weight:500;color:var(--fg);line-height:1.4;min-width:0}
.ft-row__title-text{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.ft-row__meta{display:flex;align-items:center;gap:5px;flex-wrap:wrap;min-height:0}
.ft-row__summary{font-size:11px;line-height:1.4;color:var(--fg-2);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:400px}
.ft-row__notes{font-size:10.5px;line-height:1.4;color:var(--muted);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:260px}
.ft-badges{display:inline-flex;gap:3px;align-items:center}
.ft-badge{display:inline-flex;align-items:center;padding:1px 5px;border-radius:99px;font-size:9.5px;border:1px solid var(--border);background:var(--panel);color:var(--fg-2);line-height:1.3}
.ft-badge--priority-high{background:var(--danger-soft);border-color:transparent;color:var(--danger)}
.ft-badge--priority-medium{background:var(--warning-soft);border-color:transparent;color:var(--warning)}
.ft-badge--priority-low{background:var(--success-soft);border-color:transparent;color:var(--success)}
.ft-btn--danger{background:var(--danger-soft);border-color:var(--danger);color:var(--danger)}
.ft-btn--danger:hover{background:var(--danger);border-color:var(--danger);color:#fff}
.ft-editor{display:none;padding:10px;border:1px solid var(--border);border-radius:8px;background:var(--bg-2);margin-top:4px}
.ft-editor--open{display:block}
.ft-editor__title{font-size:12px;font-weight:600;color:var(--fg);margin-bottom:8px}
.ft-form-row{display:flex;flex-direction:column;gap:3px;margin-bottom:8px}
.ft-form-row label{font-size:11px;font-weight:500;color:var(--fg-2)}
.ft-form-row input,.ft-form-row textarea,.ft-form-row select{width:100%;padding:5px 8px;border-radius:6px;border:1px solid var(--border);background:var(--panel);color:var(--fg);font-size:12px}
.ft-form-row textarea{min-height:50px;resize:vertical}
.ft-form-row input:focus,.ft-form-row textarea:focus,.ft-form-row select:focus{outline:none;border-color:var(--accent)}
.ft-grid{display:grid;grid-template-columns:1fr 140px 110px;gap:8px}
.ft-error{display:none;color:var(--danger);font-size:11px;margin-top:-2px;margin-bottom:8px}
.ft-editor__actions{display:flex;justify-content:flex-end;gap:6px;flex-wrap:wrap}
@media(max-width:768px){#ft-modal .modal{width:min(94vw,94vw);max-width:94vw;max-height:90vh}.ft-topbar{flex-wrap:wrap}.ft-tabs{flex-wrap:wrap}.ft-grid{grid-template-columns:1fr}.ft-row{flex-wrap:wrap;gap:4px}.ft-row__meta{order:10;width:100%}}
`;
document.head.appendChild(style);

const composer = document.querySelector('.composer');
const actionHost = document.getElementById('footer-actions');
if (!composer || !composer.parentNode) {
  return;
}

let trackerFeatures = [];
let editingFeatureId = null;
let loading = false;
let loadError = '';
let loadingFeatures = false;
let activeTab = 'wishlist';
let searchQuery = '';
let hasUserSelectedTab = false;
let searchTimer = null;

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
  editingFeatureId = null;
  const editor = document.getElementById('ft-editor');
  if (!editor) return;
  editor.classList.remove('ft-editor--open');
  document.getElementById('ft-title').value = '';
  document.getElementById('ft-status').value = activeTab || 'wishlist';
  document.getElementById('ft-priority').value = 'medium';
  document.getElementById('ft-summary').value = '';
  document.getElementById('ft-notes').value = '';
  document.getElementById('ft-editor-title').textContent = t('新增特性', 'Add Feature');
  document.getElementById('ft-delete').style.display = 'none';
  setFeatureTrackerError('');
}

function showFeatureEditor(feature) {
  const editor = document.getElementById('ft-editor');
  if (!editor) return;
  editingFeatureId = feature?.id || null;
  document.getElementById('ft-title').value = feature?.title || '';
  document.getElementById('ft-status').value = feature?.status || activeTab || 'wishlist';
  document.getElementById('ft-priority').value = feature?.priority || 'medium';
  document.getElementById('ft-summary').value = feature?.summary || '';
  document.getElementById('ft-notes').value = feature?.notes || '';
  document.getElementById('ft-editor-title').textContent = editingFeatureId ? t('编辑特性', 'Edit Feature') : t('新增特性', 'Add Feature');
  document.getElementById('ft-delete').style.display = editingFeatureId ? '' : 'none';
  setFeatureTrackerError('');
  editor.classList.add('ft-editor--open');
  document.getElementById('ft-title').focus();
}

function renderTabs() {
  const tabs = document.getElementById('ft-tabs');
  if (!tabs) return;
  tabs.innerHTML = '';
  FEATURE_STATUSES.forEach((status) => {
    const count = trackerFeatures.filter((f) => f.status === status).length;
    const tab = document.createElement('button');
    tab.className = 'ft-tab' + (activeTab === status ? ' ft-tab--active' : '');
    tab.type = 'button';
    tab.dataset.status = status;
    tab.innerHTML = escapeHTML(statusLabel(status)) + ' <span class="ft-tab__badge">' + count + '</span>';
    tabs.appendChild(tab);
  });
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
  let filtered = trackerFeatures;
  if (activeTab) {
    filtered = filtered.filter((f) => f.status === activeTab);
  }
  if (searchQuery) {
    const q = searchQuery.toLowerCase();
    filtered = filtered.filter((f) =>
      f.title.toLowerCase().includes(q) ||
      (f.summary && f.summary.toLowerCase().includes(q)) ||
      (f.notes && f.notes.toLowerCase().includes(q))
    );
  }
  if (!filtered.length) {
    const empty = document.createElement('div');
    empty.className = 'ft-empty';
    empty.textContent = searchQuery
      ? t('没有匹配的特性条目。', 'No matching features found.')
      : t('当前标签下没有特性条目。', 'No features in this status.');
    list.appendChild(empty);
    return;
  }
  filtered.forEach((feature) => {
    const row = document.createElement('div');
    row.className = 'ft-row';
    const p = FEATURE_PRIORITIES.includes(feature.priority) ? feature.priority : 'medium';
    var metaHTML = '';
    if (feature.summary) {
      metaHTML += '<span class="ft-row__summary">' + escapeHTML(feature.summary) + '</span>';
    }
    if (feature.notes) {
      metaHTML += '<span class="ft-row__notes">' + escapeHTML(feature.notes) + '</span>';
    }
    row.dataset.id = feature.id;
    row.innerHTML =
      '<div class="ft-row__main">' +
        '<div class="ft-row__title">' +
          '<span class="ft-badge ft-badge--priority-' + escapeHTML(p) + '">' + escapeHTML(priorityLabel(p)) + '</span>' +
          '<span class="ft-row__title-text">' + escapeHTML(feature.title) + '</span>' +
        '</div>' +
        '<div class="ft-row__meta">' + metaHTML + '</div>' +
      '</div>';
    list.appendChild(row);
  });
}

function persistFeatures(nextFeatures, onSuccess) {
  if (loading) return;
  loading = true;
  const saveButton = document.getElementById('ft-save');
  if (saveButton) saveButton.disabled = true;
  setFeatureTrackerError('');
  fetch('/project-features', {
    method: 'POST',
    headers: {'Content-Type': 'application/json', 'X-Requested-With': 'XMLHttpRequest'},
    body: JSON.stringify({features: nextFeatures})
  }).then((response) => {
    if (!response.ok) throw new Error('save failed');
    return response.json();
  }).then((snapshot) => {
    trackerFeatures = cloneFeatures(snapshot.features);
    renderTabs();
    renderFeatureSections();
    if (typeof onSuccess === 'function') onSuccess();
  }).catch((error) => {
    console.error('save failed', error);
    setFeatureTrackerError(t('保存失败，请重试。', 'Save failed, please retry.'));
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
  const safeStatus = FEATURE_STATUSES.includes(status) ? status : 'wishlist';
  const safePriority = FEATURE_PRIORITIES.includes(priority) ? priority : 'medium';
  const nextFeatures = cloneFeatures(trackerFeatures);
  const next = {id: editingFeatureId, title, status: safeStatus, priority: safePriority, summary, notes};
  const index = nextFeatures.findIndex((feature) => feature.id === editingFeatureId);
  if (index >= 0) {
    nextFeatures[index] = next;
  } else {
    nextFeatures.push(next);
  }
  persistFeatures(nextFeatures, () => {
    activeTab = safeStatus;
    renderTabs();
    renderFeatureSections();
    hideFeatureEditor();
  });
}

function deleteFeature(id) {
  if (!id) return;
  const feature = trackerFeatures.find((item) => item.id === id);
  if (!feature) return;
  const nextFeatures = cloneFeatures(trackerFeatures).filter((item) => item.id !== id);
  persistFeatures(nextFeatures, () => {
    if (editingFeatureId === id) hideFeatureEditor();
  });
}

function handleFeatureListClick(event) {
  const row = event.target.closest('.ft-row');
  if (!row) return;
  const id = row.dataset.id;
  if (!id) return;
  const feature = trackerFeatures.find((item) => item.id === id);
  if (feature) showFeatureEditor(feature);
}

function loadFeatureTracker() {
  if (loadingFeatures) return;
  loadingFeatures = true;
  loadError = '';
  trackerFeatures = [];
  renderTabs();
  renderFeatureSections();
  return fetch('/project-features')
    .then((response) => {
      if (!response.ok) throw new Error('load failed');
      return response.json();
    })
    .then((snapshot) => {
      loadingFeatures = false;
      loadError = '';
      trackerFeatures = cloneFeatures(snapshot.features);
      // auto-select first non-empty tab only on initial load
      if (!hasUserSelectedTab) {
        const firstNonEmpty = FEATURE_STATUSES.find((s) => trackerFeatures.some((f) => f.status === s));
        if (firstNonEmpty) activeTab = firstNonEmpty;
      }
      renderTabs();
      renderFeatureSections();
    })
    .catch(() => {
      loadingFeatures = false;
      loadError = t('加载项目特性失败，请稍后重试。', 'Failed to load project features. Please try again.');
      renderTabs();
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
      '<div class="ft-topbar">' +
        '<div class="ft-search">' +
          '<svg class="ft-search__icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>' +
          '<input type="text" class="ft-search__input" id="ft-search" placeholder="' + t('搜索标题/描述...', 'Search title/notes...') + '" />' +
        '</div>' +
        '<button type="button" class="ft-btn ft-btn--primary" id="ft-add" title="' + t('新增特性', 'Add Feature') + '">' +
          '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>' +
        '</button>' +
      '</div>' +
      '<div class="ft-tabs" id="ft-tabs"></div>' +
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
          '<button type="button" class="ft-btn ft-btn--danger" id="ft-delete" style="margin-right:auto">' + t('删除此条目', 'Delete') + '</button>' +
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
  document.getElementById('ft-delete').onclick = function() {
    if (editingFeatureId && window.confirm(t('确认删除这个特性条目？', 'Delete this feature entry?'))) {
      deleteFeature(editingFeatureId);
    }
  };
  document.getElementById('ft-add').onclick = () => showFeatureEditor(null);
  document.getElementById('ft-list').onclick = handleFeatureListClick;
  document.getElementById('ft-tabs').onclick = (event) => {
    const tab = event.target.closest('.ft-tab');
    if (!tab) return;
    if (searchTimer) clearTimeout(searchTimer);
    hasUserSelectedTab = true;
    activeTab = tab.dataset.status;
    renderTabs();
    renderFeatureSections();
  };
  document.getElementById('ft-search').addEventListener('input', (event) => {
    searchQuery = event.target.value;
    if (searchTimer) clearTimeout(searchTimer);
    searchTimer = setTimeout(renderFeatureSections, 150);
  });
  overlay.onclick = (event) => {
    if (event.target === overlay) closeFeatureTracker();
  };

  return overlay;
}

function closeFeatureTracker() {
  if (searchTimer) { clearTimeout(searchTimer); searchTimer = null; }
  const modal = document.getElementById('ft-modal');
  if (modal) modal.style.display = 'none';
  hideFeatureEditor();
}

function openFeatureTracker() {
  const modal = ensureFeatureTrackerModal();
  if (modal.style.display !== 'none') return; // already open
  modal.style.display = 'flex';
  hasUserSelectedTab = false;
  searchQuery = '';
  const searchInput = document.getElementById('ft-search');
  if (searchInput) searchInput.value = '';
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
