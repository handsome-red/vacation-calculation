import { initStorage } from './storage.js';
import { renderQuotas } from './quota.js';

const dataEl = document.getElementById('app-data');
const APP = JSON.parse(dataEl.textContent);

console.log('APP =', APP);

initStorage(APP);

const root = document.getElementById('quota-app');
renderQuotas(root, APP);

window.addEventListener('storage', (e) => {
  if (e.key === `vacation_reserved:${APP.userId}`) {
    renderQuotas(root, APP);
  }
});