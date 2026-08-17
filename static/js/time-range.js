(function () {
	'use strict';

	function parseTimeToMinutes(value) {
		if (!value) return null;
		var match = String(value).match(/^(\d{1,2}):(\d{2})$/);
		if (!match) return null;
		return parseInt(match[1], 10) * 60 + parseInt(match[2], 10);
	}

	function formatDurationLabel(minutes) {
		if (minutes < 60) return minutes + ' min';
		var h = Math.floor(minutes / 60);
		var m = minutes % 60;
		return m === 0 ? h + ' hr' : h + ' hr ' + m + ' min';
	}

	function validateTimeRange(opts) {
		opts = opts || {};
		var startInput = document.getElementById(opts.startId || 'start_time');
		var endInput = document.getElementById(opts.endId || 'end_time');
		if (!startInput || !endInput) {
			return { ok: true };
		}
		if (!startInput.value || !endInput.value) {
			return { ok: false, message: 'Please set start and end times' };
		}
		var start = parseTimeToMinutes(startInput.value);
		var end = parseTimeToMinutes(endInput.value);
		if (start === null || end === null) {
			return { ok: false, message: 'Invalid time format' };
		}
		if (end <= start) {
			return { ok: false, message: 'End time must be after start time' };
		}
		return { ok: true };
	}

	function showTimeRangeError(message) {
		var preview = document.getElementById('durationPreview');
		if (preview) {
			preview.hidden = false;
			preview.textContent = message;
			preview.classList.add('is-invalid');
		}
	}

	function initTimeRangePreview(opts) {
		opts = opts || {};
		var startInput = document.getElementById(opts.startId || 'start_time');
		var endInput = document.getElementById(opts.endId || 'end_time');
		var preview = document.getElementById(opts.previewId || 'durationPreview');
		if (!startInput || !endInput || !preview) return;

		function update() {
			var start = parseTimeToMinutes(startInput.value);
			var end = parseTimeToMinutes(endInput.value);
			if (start === null || end === null) {
				preview.textContent = '';
				preview.hidden = true;
				preview.classList.remove('is-invalid');
				return;
			}
			if (end <= start) {
				preview.hidden = false;
				preview.textContent = 'End time must be after start time';
				preview.classList.add('is-invalid');
				return;
			}
			preview.hidden = false;
			preview.textContent = 'Duration: ' + formatDurationLabel(end - start);
			preview.classList.remove('is-invalid');
		}

		startInput.addEventListener('change', update);
		endInput.addEventListener('change', update);

		var form = startInput.closest('form');
		if (form) {
			form.addEventListener('reset', function () {
				window.requestAnimationFrame(update);
			});
		}

		update();
	}

	function attachTimeRangeFormValidation() {
		document.querySelectorAll('form').forEach(function (form) {
			if (!form.querySelector('[name="start_time"]') || !form.querySelector('[name="end_time"]')) {
				return;
			}
			if (form.dataset.timeRangeValidation === 'true') {
				return;
			}
			form.dataset.timeRangeValidation = 'true';
			form.addEventListener('submit', function (e) {
				var result = validateTimeRange();
				if (!result.ok) {
					e.preventDefault();
					e.stopImmediatePropagation();
					showTimeRangeError(result.message);
					alert(result.message);
				}
			}, true);
		});
	}

	window.validateTimeRange = validateTimeRange;
	window.initTimeRangePreview = initTimeRangePreview;
	window.attachTimeRangeFormValidation = attachTimeRangeFormValidation;
})();
