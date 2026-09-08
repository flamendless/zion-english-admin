(function () {
	'use strict';

	function parseDateParts(dateStr) {
		var parts = (dateStr || '').split('-').map(Number);
		if (parts.length !== 3) return null;
		return { y: parts[0], m: parts[1], d: parts[2] };
	}

	function formatISO(y, m, d) {
		return y + '-' + String(m).padStart(2, '0') + '-' + String(d).padStart(2, '0');
	}

	function formatDisplay(dateStr) {
		var parts = parseDateParts(dateStr);
		if (!parts) return dateStr;
		var dt = new Date(parts.y, parts.m - 1, parts.d);
		return dt.toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' });
	}

	function initScheduleRepeatCalendar(root, overrides) {
		if (!root || root.dataset.scheduleRepeatCalendarInit === 'true') return null;
		root.dataset.scheduleRepeatCalendarInit = 'true';

		overrides = overrides || {};
		var todayPHT = overrides.todayPHT || root.dataset.todayPht || '';
		var minDates = parseInt(overrides.minDates || root.dataset.minDates || '2', 10);
		var inputName = overrides.inputName || root.dataset.inputName || 'scheduled_dates';
		var submitMessage = overrides.submitMessage || root.dataset.submitMessage || 'Select at least two future dates, or use Schedule a Class for a single date.';
		var readonly = root.dataset.readonly === 'true' || overrides.readonly === true;

		var grid = root.querySelector('[data-cal-grid]');
		var monthLabel = root.querySelector('[data-cal-month-label]');
		var prevBtn = root.querySelector('[data-cal-prev]');
		var nextBtn = root.querySelector('[data-cal-next]');
		var selectedWrap = root.querySelector('[data-cal-selected]');
		var inputsWrap = root.querySelector('[data-cal-inputs]');
		var form = root.closest('form');
		var selected = new Set();
		var viewYear = 0;
		var viewMonth = 0;

		var initialRaw = root.dataset.initialDates || '';
		if (initialRaw) {
			initialRaw.split(',').forEach(function (date) {
				date = date.trim();
				if (date) selected.add(date);
			});
		}

		function initViewFromToday() {
			var parts = parseDateParts(todayPHT);
			if (!parts) {
				var now = new Date();
				viewYear = now.getFullYear();
				viewMonth = now.getMonth() + 1;
				return;
			}
			viewYear = parts.y;
			viewMonth = parts.m;
		}

		function syncHiddenInputs() {
			if (!inputsWrap) return;
			inputsWrap.innerHTML = '';
			Array.from(selected).sort().forEach(function (date) {
				var input = document.createElement('input');
				input.type = 'hidden';
				input.name = inputName;
				input.value = date;
				inputsWrap.appendChild(input);
			});
		}

		function syncChips() {
			if (!selectedWrap) return;
			selectedWrap.innerHTML = '';
			if (!selected.size) {
				selectedWrap.textContent = 'No dates selected yet.';
				return;
			}
			Array.from(selected).sort().forEach(function (date) {
				var chip = document.createElement('span');
				chip.className = 'schedule-repeat-date-chip';
				chip.textContent = formatDisplay(date);
				selectedWrap.appendChild(chip);
			});
		}

		function syncActiveDateField() {
			if (!form) return;
			var activeInput = form.querySelector('[data-series-active-date]');
			if (!activeInput) return;
			var original = activeInput.getAttribute('data-original-active-date');
			if (!original) {
				original = activeInput.value;
				activeInput.setAttribute('data-original-active-date', original);
			}
			if (selected.has(original)) {
				activeInput.value = original;
				return;
			}
			var initial = (root.dataset.initialDates || '').split(',').map(function (d) {
				return d.trim();
			}).filter(Boolean);
			var added = null;
			Array.from(selected).forEach(function (date) {
				if (initial.indexOf(date) === -1) added = date;
			});
			if (added) activeInput.value = added;
		}

		function renderCalendar() {
			if (!grid || !monthLabel) return;
			monthLabel.textContent = new Date(viewYear, viewMonth - 1, 1).toLocaleDateString(undefined, { month: 'long', year: 'numeric' });
			grid.innerHTML = '';
			['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].forEach(function (label) {
				var head = document.createElement('div');
				head.className = 'schedule-repeat-weekday';
				head.textContent = label;
				grid.appendChild(head);
			});
			var firstDay = new Date(viewYear, viewMonth - 1, 1).getDay();
			var daysInMonth = new Date(viewYear, viewMonth, 0).getDate();
			var i;
			for (i = 0; i < firstDay; i++) {
				var empty = document.createElement('div');
				empty.className = 'schedule-repeat-day is-empty';
				grid.appendChild(empty);
			}
			for (var day = 1; day <= daysInMonth; day++) {
				var dateStr = formatISO(viewYear, viewMonth, day);
				var btn = document.createElement('button');
				btn.type = 'button';
				btn.className = 'schedule-repeat-day';
				btn.textContent = String(day);
				btn.setAttribute('aria-label', formatDisplay(dateStr));
				if (dateStr < todayPHT) {
					btn.classList.add('is-disabled');
					btn.disabled = true;
				}
				if (selected.has(dateStr)) {
					btn.classList.add('is-selected');
				}
				if (!readonly) {
					btn.addEventListener('click', function (picked) {
						return function () {
							if (picked < todayPHT) return;
							if (selected.has(picked)) {
								selected.delete(picked);
							} else {
								selected.add(picked);
							}
							syncHiddenInputs();
							syncChips();
							syncActiveDateField();
							renderCalendar();
						};
					}(dateStr));
				} else {
					btn.disabled = true;
					btn.classList.add('is-disabled');
				}
				grid.appendChild(btn);
			}
		}

		function toggleDate(dateStr) {
			if (readonly || dateStr < todayPHT) return;
			if (selected.has(dateStr)) {
				selected.delete(dateStr);
			} else {
				selected.add(dateStr);
			}
			syncHiddenInputs();
			syncChips();
			syncActiveDateField();
			renderCalendar();
		}

		if (prevBtn) {
			prevBtn.addEventListener('click', function () {
				viewMonth -= 1;
				if (viewMonth < 1) {
					viewMonth = 12;
					viewYear -= 1;
				}
				renderCalendar();
			});
		}
		if (nextBtn) {
			nextBtn.addEventListener('click', function () {
				viewMonth += 1;
				if (viewMonth > 12) {
					viewMonth = 1;
					viewYear += 1;
				}
				renderCalendar();
			});
		}

		if (form) {
			form.addEventListener('submit', function (e) {
				if (readonly) return;
				if (selected.size < minDates) {
					e.preventDefault();
					alert(submitMessage);
				}
			});
		}

		var api = {
			setReadonly: function (flag) {
				readonly = !!flag;
				root.dataset.readonly = readonly ? 'true' : 'false';
				if (prevBtn) prevBtn.disabled = readonly;
				if (nextBtn) nextBtn.disabled = readonly;
				renderCalendar();
			},
			getSelected: function () {
				return Array.from(selected).sort();
			},
			setSelected: function (dates) {
				selected.clear();
				(dates || []).forEach(function (date) {
					if (date) selected.add(date);
				});
				syncHiddenInputs();
				syncChips();
				syncActiveDateField();
				renderCalendar();
			},
			toggleDate: toggleDate
		};

		root._scheduleRepeatCalendar = api;
		initViewFromToday();
		syncHiddenInputs();
		syncChips();
		syncActiveDateField();
		renderCalendar();
		return api;
	}

	function bootScheduleRepeatCalendars() {
		document.querySelectorAll('[data-schedule-repeat-calendar]').forEach(function (root) {
			initScheduleRepeatCalendar(root);
		});
	}

	window.initScheduleRepeatCalendar = initScheduleRepeatCalendar;
	window.bootScheduleRepeatCalendars = bootScheduleRepeatCalendars;

	if (document.readyState === 'loading') {
		document.addEventListener('DOMContentLoaded', bootScheduleRepeatCalendars);
	} else {
		bootScheduleRepeatCalendars();
	}
})();
