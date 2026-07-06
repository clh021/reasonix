(function() {
'use strict';

const lang = (document.documentElement.lang || 'en').startsWith('zh') ? 'zh' : 'en';
const t = (zh, en) => (lang === 'zh' ? zh : en);

const style = document.createElement('style');
style.textContent =
`.ra-menu{position:fixed;z-index:210;display:none;min-width:220px;max-width:min(280px,calc(100vw - 24px));padding:6px;background:var(--panel);border:1px solid var(--border-strong);border-radius:12px;box-shadow:var(--shadow-lg)}
.ra-menu--open{display:block}
.ra-modal{position:fixed;inset:0;z-index:220;display:none;align-items:center;justify-content:center;background:oklch(0% 0 0/.55)}
.ra-modal--open{display:flex}
.ra-modal__card{display:flex;flex-direction:column;width:min(760px,92vw);max-width:92vw;max-height:min(88vh,900px);background:var(--panel);border:1px solid var(--border-strong);border-radius:12px;box-shadow:var(--shadow-lg);overflow:hidden}
.ra-modal__head{display:flex;align-items:center;gap:8px;padding:14px 16px;border-bottom:1px solid var(--border);font-family:var(--heading);font-weight:600;font-size:15px}
.ra-modal__close{margin-left:auto;width:24px;height:24px;display:flex;align-items:center;justify-content:center;border-radius:6px;color:var(--muted-2);cursor:pointer}
.ra-modal__close:hover{background:var(--card-hover);color:var(--fg)}
.ra-modal__body{display:flex;flex-direction:column;gap:12px;padding:14px 16px;min-height:0;overflow:auto}
.ra-modal__meta{display:flex;flex-wrap:wrap;gap:8px}
.ra-badge{display:inline-flex;align-items:center;padding:2px 7px;border-radius:999px;font-size:10.5px;border:1px solid var(--border);background:var(--bg-2);color:var(--fg-2)}
.ra-output{padding:12px;border:1px solid var(--border);border-radius:10px;background:var(--bg-2);font-family:var(--mono);font-size:12px;line-height:1.55;color:var(--fg);white-space:pre-wrap;word-break:break-word}
.ra-error{padding:12px;border:1px solid var(--danger);border-radius:10px;background:var(--danger-soft);font-size:12.5px;line-height:1.55;color:var(--danger)}
.ra-menu__head{padding:8px 10px 6px;font-size:11px;font-weight:700;letter-spacing:.04em;text-transform:uppercase;color:var(--muted)}
.ra-menu__item{display:flex;align-items:flex-start;gap:10px;width:100%;padding:10px;border:none;border-radius:9px;background:none;color:inherit;text-align:left;cursor:pointer;transition:background .15s ease}
.ra-menu__item:hover{background:var(--card-hover)}
.ra-menu__icon{width:16px;height:16px;flex-shrink:0;color:var(--accent);margin-top:1px}
.ra-menu__body{display:flex;flex-direction:column;gap:2px;min-width:0}
.ra-menu__title{font-size:12.5px;font-weight:600;color:var(--fg)}
.ra-menu__desc{font-size:11.5px;line-height:1.45;color:var(--fg-2)}
@media(max-width:768px){.ra-menu{left:12px!important;right:12px;top:auto!important;bottom:108px;max-width:none;min-width:0}.ra-modal__card{width:min(94vw,94vw);max-width:94vw;max-height:90vh}}
`;
document.head.appendChild(style);

const composer = document.querySelector('.composer');
const actionHost = document.getElementById('footer-actions');
if (!composer || !composer.parentNode) {
  return;
}

let runningAction = false;

function placeMenu(menu, anchor) {
  const rect = anchor.getBoundingClientRect();
  menu.style.left = '0px';
  menu.style.top = '0px';
  menu.classList.add('ra-menu--open');
  const menuRect = menu.getBoundingClientRect();
  const gap = 8;
  const maxLeft = Math.max(12, window.innerWidth - menuRect.width - 12);
  const left = Math.min(Math.max(12, rect.left), maxLeft);
  const preferAbove = rect.top - menuRect.height - gap;
  const top = preferAbove >= 12 ? preferAbove : Math.min(window.innerHeight - menuRect.height - 12, rect.bottom + gap);
  menu.style.left = left + 'px';
  menu.style.top = Math.max(12, top) + 'px';
}

function closeMenu() {
  const menu = document.getElementById('ra-menu');
  if (!menu) return;
  menu.classList.remove('ra-menu--open');
}

function closeModal() {
  const modal = document.getElementById('ra-modal');
  if (!modal) return;
  modal.classList.remove('ra-modal--open');
}

function ensureModal() {
  let modal = document.getElementById('ra-modal');
  if (modal) return modal;
  modal = document.createElement('div');
  modal.id = 'ra-modal';
  modal.className = 'ra-modal';
  modal.innerHTML =
    '<div class="ra-modal__card">' +
      '<div class="ra-modal__head">' +
        '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 17l6-5-6-5"/><path d="M12 19h8"/><path d="M12 5h8"/></svg>' +
        '<span id="ra-modal-title">' + t('仓库操作', 'Repository Actions') + '</span>' +
        '<span class="ra-modal__close" id="ra-modal-close">&times;</span>' +
      '</div>' +
      '<div class="ra-modal__body" id="ra-modal-body"></div>' +
    '</div>';
  modal.onclick = (event) => {
    if (event.target === modal) closeModal();
  };
  document.body.appendChild(modal);
  document.getElementById('ra-modal-close').onclick = closeModal;
  window.addEventListener('keydown', (event) => {
    if (event.key === 'Escape') closeModal();
  });
  return modal;
}

function renderResponse(result) {
  const modal = ensureModal();
  const body = document.getElementById('ra-modal-body');
  const title = document.getElementById('ra-modal-title');
  if (!body || !title) return;

  title.textContent = result.action === 'status' ? 'Git Status' : 'Git Push';
  const meta = [];
  if (result.repoRoot) meta.push('<span class="ra-badge">' + escapeHTML(result.repoRoot) + '</span>');
  if (result.branch) meta.push('<span class="ra-badge">' + escapeHTML(result.detached ? 'HEAD@' + result.branch : result.branch) + '</span>');

  body.innerHTML =
    (meta.length ? '<div class="ra-modal__meta">' + meta.join('') + '</div>' : '') +
    (result.error
      ? '<div class="ra-error">' + escapeHTML(result.error) + '</div>'
      : '<div class="ra-output">' + escapeHTML(result.output || '') + '</div>');
  modal.classList.add('ra-modal--open');
}

function escapeHTML(value) {
  return String(value)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function runAction(kind) {
  if (runningAction) return;
  runningAction = true;
  closeMenu();
  fetch('/repo-action', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({action: kind})
  }).then((response) => {
    if (!response.ok) throw new Error('request failed');
    return response.json();
  }).then((result) => {
    renderResponse(result || {action: kind, error: t('空响应', 'Empty response')});
  }).catch((error) => {
    renderResponse({action: kind, error: t('仓库操作失败：', 'Repository action failed: ') + error.message});
  }).finally(() => {
    runningAction = false;
  });
}

function ensureMenu() {
  let menu = document.getElementById('ra-menu');
  if (menu) return menu;

  menu = document.createElement('div');
  menu.id = 'ra-menu';
  menu.className = 'ra-menu';
  menu.innerHTML =
    '<div class="ra-menu__head">' + t('仓库操作', 'Repository Actions') + '</div>' +
    '<button type="button" class="ra-menu__item" data-kind="status">' +
      '<svg class="ra-menu__icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 3v18h18"/><path d="M7 14l3-3 4 4 5-8"/></svg>' +
      '<span class="ra-menu__body">' +
        '<span class="ra-menu__title">Git Status</span>' +
        '<span class="ra-menu__desc">' + t('直接在当前项目目录执行 git status 并显示结果。', 'Run git status in the current project and show the result.') + '</span>' +
      '</span>' +
    '</button>' +
    '<button type="button" class="ra-menu__item" data-kind="push">' +
      '<svg class="ra-menu__icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 19V5"/><path d="M5 12l7-7 7 7"/></svg>' +
      '<span class="ra-menu__body">' +
        '<span class="ra-menu__title">Git Push</span>' +
        '<span class="ra-menu__desc">' + t('保留仓库推送入口，当前未启用实际执行。', 'Reserved push entry; real execution is not enabled yet.') + '</span>' +
      '</span>' +
    '</button>';
  menu.onclick = (event) => {
    const item = event.target.closest('[data-kind]');
    if (!item) return;
    runAction(item.dataset.kind);
  };
  document.body.appendChild(menu);

  document.addEventListener('click', (event) => {
    if (!menu.classList.contains('ra-menu--open')) return;
    if (event.target.closest('#ra-menu') || event.target.closest('.ra-launch')) return;
    closeMenu();
  });
  window.addEventListener('resize', closeMenu);
  window.addEventListener('keydown', (event) => {
    if (event.key === 'Escape') closeMenu();
  });
  return menu;
}

function toggleMenu(event) {
  const menu = ensureMenu();
  if (menu.classList.contains('ra-menu--open')) {
    closeMenu();
    return;
  }
  placeMenu(menu, event.currentTarget);
}

const launch = document.createElement('button');
launch.type = 'button';
launch.className = 'footer-action ra-launch';
launch.style.order = '30';
launch.innerHTML =
  '<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 17l6-5-6-5"/><path d="M12 19h8"/><path d="M12 5h8"/></svg>' +
  '<span>' + t('仓库', 'Repo') + '</span>';
launch.onclick = toggleMenu;

if (actionHost) {
  actionHost.appendChild(launch);
} else {
  composer.appendChild(launch);
}
})();
