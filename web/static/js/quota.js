const TYPE_LABELS = {
  base: 'Ежегодный отпуск',
  seniority: 'За выслугу лет',
  irregular: 'За ненормированный служебный день',
};

function createQuotaBar(year, type, quota) {
  const total = quota.total;
  const blocked = quota.blocked;
  const reserved = quota.reserved || 0;
  const free = Math.max(0, total - blocked - reserved);

  const pct = (n) => total === 0 ? 0 : (n / total) * 100;

  const wrap = document.createElement('div');
  wrap.className = 'quota';

  const header = document.createElement('div');
  header.className = 'quota__header';

  const title = document.createElement('span');
  title.className = 'quota__title';
  title.textContent = TYPE_LABELS[type];

  const totalLabel = document.createElement('span');
  totalLabel.className = 'quota__total';
  totalLabel.textContent = `${total} дн.`;

  header.append(title, totalLabel);

  const bar = document.createElement('div');
  bar.className = 'quota__bar';

  const segBlocked = document.createElement('div');
  segBlocked.className = 'quota__seg quota__seg--blocked';
  segBlocked.style.width = `${pct(blocked)}%`;

  const segReserved = document.createElement('div');
  segReserved.className = 'quota__seg quota__seg--reserved';
  segReserved.style.width = `${pct(reserved)}%`;

  const segFree = document.createElement('div');
  segFree.className = 'quota__seg quota__seg--free';
  segFree.style.width = `${pct(free)}%`;

  bar.append(segBlocked, segReserved, segFree);

  const legend = document.createElement('div');
  legend.className = 'quota__legend';
  legend.innerHTML = `
    <span><i class="dot dot--blocked"></i>списано ${blocked}</span>
    <span><i class="dot dot--reserved"></i>в заявке ${reserved}</span>
    <span><i class="dot dot--free"></i>свободно ${free}</span>
  `;

  wrap.append(header, bar, legend);
  return wrap;
}

function createYearCard(stat) {
  const card = document.createElement('section');
  card.className = 'year-card';

  const h2 = document.createElement('h2');
  h2.className = 'year-card__title';
  h2.textContent = stat.year;
  card.append(h2);

  const bars = document.createElement('div');
  bars.className = 'year-card__bars';

  // Порядок совпадает с порядком списания внутри рабочего года.
  for (const type of ['base', 'irregular', 'seniority']) {
    bars.append(createQuotaBar(stat.year, type, stat[type]));
  }

  card.append(bars);
  return card;
}

export function renderQuotas(rootEl, app) {
  rootEl.replaceChildren();

  if (!app.yearStats || app.yearStats.length === 0) {
    const p = document.createElement('p');
    p.textContent = 'Нет данных о рабочих годах';
    rootEl.append(p);
    return;
  }

  for (const stat of app.yearStats) {
    rootEl.append(createYearCard(stat));
  }
}