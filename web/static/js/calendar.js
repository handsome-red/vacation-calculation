(function () {
	const data = JSON.parse(document.getElementById('calendar-data').textContent);
	const year = data.year;
	const holidayMap = new Map(data.holidays.map(h => [h.date, h.name]));

	const calendarEl = document.getElementById('calendar');
	const listEl = document.getElementById('holidays-list');

	const monthNames = [
		'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
		'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
	];

	const weekDays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];

	function pad(n) { return String(n).padStart(2, '0'); }

	function formatDate(d) {
		return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
	}

	function renderMonth(month) {
		const first = new Date(year, month, 1);
		const last = new Date(year, month + 1, 0);

		// Понедельник — первый день недели
		let startWeekDay = first.getDay();
		if (startWeekDay === 0) startWeekDay = 7;
		startWeekDay -= 1;

		const table = document.createElement('table');
		table.className = 'month';

		const caption = document.createElement('caption');
		caption.textContent = monthNames[month];
		table.appendChild(caption);

		const thead = document.createElement('thead');
		const headRow = document.createElement('tr');
		for (const wd of weekDays) {
			const th = document.createElement('th');
			th.textContent = wd;
			headRow.appendChild(th);
		}
		thead.appendChild(headRow);
		table.appendChild(thead);

		const tbody = document.createElement('tbody');
		let row = document.createElement('tr');

		for (let i = 0; i < startWeekDay; i++) {
			row.appendChild(document.createElement('td'));
		}

		for (let day = 1; day <= last.getDate(); day++) {
			const d = new Date(year, month, day);
			const iso = formatDate(d);

			const td = document.createElement('td');
			td.textContent = day;

			if (holidayMap.has(iso)) {
				td.classList.add('holiday');
				td.title = holidayMap.get(iso);
			}

			row.appendChild(td);

			if (row.children.length === 7) {
				tbody.appendChild(row);
				row = document.createElement('tr');
			}
		}

		if (row.children.length > 0) {
			while (row.children.length < 7) {
				row.appendChild(document.createElement('td'));
			}
			tbody.appendChild(row);
		}

		table.appendChild(tbody);
		return table;
	}

	for (let m = 0; m < 12; m++) {
		calendarEl.appendChild(renderMonth(m));
	}

	// Список праздников
	const sorted = [...data.holidays].sort((a, b) => a.date.localeCompare(b.date));
	for (const h of sorted) {
		const li = document.createElement('li');
		const [y, m, d] = h.date.split('-');
		li.textContent = `${d}.${m}.${y} — ${h.name}`;
		listEl.appendChild(li);
	}
})();