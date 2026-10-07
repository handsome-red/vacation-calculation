document.addEventListener('DOMContentLoaded', () => {
  document.querySelectorAll('.row[data-toggle]').forEach(row => {
    const btn = row.querySelector('.chevron');
    if (!btn) return;

    btn.addEventListener('click', () => {
      const details = document.getElementById(btn.getAttribute('aria-controls'));
      if (!details) return;

      const willOpen = details.hidden;
      details.hidden = !willOpen;
      btn.setAttribute('aria-expanded', String(willOpen));
      btn.textContent = willOpen ? '▾' : '▸';
    });
  });
});