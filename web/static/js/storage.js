let APP = null;

export function initStorage(app) {
    APP = app;
}

function key() {
    return `vacation_reserved:${APP.userId}`;
}

export function loadAll() {
    try {
        const raw = localStorage.getItem(key());
        return raw ? JSON.parse(raw) : {};
    } catch (e){
        console.warn('localStorage недоступен', e);
        return {};
    }
}

export function saveAll(state) {
    try {
        localStorage.setItem(key(), JSON.stringify(state))
    } catch (e) {
        console.warn('localStorage недоступен', e);
    }
}

export function reservedDays(year, type) {
  const all = loadAll();
  const items = all[year]?.[type]?.items ?? [];
  return items.reduce((sum, it) => sum + it.days, 0);
}

export function freeDays(year, type) {
  const srv = APP.yearStats.find(y => y.year === year)?.[type];
  if (!srv) return 0;
  return Math.max(0, srv.total - srv.blocked - reservedDays(year, type));
}