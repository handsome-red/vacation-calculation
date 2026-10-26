import { initStorage } from './storage.js';
import { renderQuotas } from './quota.js';

const dataEl = document.getElementById('app-data');
const APP = JSON.parse(dataEl.textContent);
window.APP = APP;

console.log('APP =', APP);

initStorage(APP);

const root = document.getElementById('quota-app');
renderQuotas(root, APP);

function allocateDays(yearStats, days) {
    const stats = yearStats.map(y => ({
        year: y.year,
        base:      { total: y.base.total,      blocked: y.base.blocked,      reserved: 0 },
        seniority: { total: y.seniority.total, blocked: y.seniority.blocked, reserved: 0 },
        irregular: { total: y.irregular.total, blocked: y.irregular.blocked, reserved: 0 },
    }));

    stats.sort((a, b) => yearStart(a.year) - yearStart(b.year));

    let left = days;
    for (const y of stats) {
        for (const kind of ["base", "seniority", "irregular"]) {
            if (left <= 0) break;
            const avail = y[kind].total - y[kind].blocked - y[kind].reserved;;
            if (avail <= 0) continue;
            const take = Math.min(avail, left);
            y[kind].reserved += take;
            left -= take;
        }
        if (left <= 0) break;
    }

    return { stats, leftover: left };
}

function yearStart(s) {
    const [from] = s.split("-");
    const [d, m, y] = from.split(".").map(Number);
    return new Date(2000 + y, m - 1, d).getTime();
}

// Реакция на изменение выбранных периодов.
window.addEventListener("periods-changed", (e) => {
    const { totalWork } = e.detail || { totalWork: 0 };
    const { stats, leftover } = allocateDays(APP.yearStats, totalWork);
    console.log("ALLOCATED first:", JSON.stringify(stats[0], null, 2));   // ← first, не last
    renderQuotas(root, { ...APP, yearStats: stats });
});

// Оставляем как было — синхронизация между вкладками (localStorage).
window.addEventListener('storage', (e) => {
    if (e.key === `vacation_reserved:${APP.userId}`) {
        renderQuotas(root, APP);
    }
});