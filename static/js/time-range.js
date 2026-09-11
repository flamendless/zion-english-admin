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

	function resolveTimeRangeTargets(opts) {
		opts = opts || {};
		var startInput = opts.startInput;
		var endInput = opts.endInput;
		var preview = opts.preview;
		if (opts.form) {
			if (!startInput) startInput = opts.form.querySelector('[name="start_time"]');
			if (!endInput) endInput = opts.form.querySelector('[name="end_time"]');
			if (!preview) preview = opts.form.querySelector('.duration-bridge');
		}
		if (!startInput) startInput = document.getElementById(opts.startId || 'start_time');
		if (!endInput) endInput = document.getElementById(opts.endId || 'end_time');
		if (!preview) preview = document.getElementById(opts.previewId || 'durationPreview');
		return { startInput: startInput, endInput: endInput, preview: preview };
	}

	function findDurationBridge(root) {
		if (!root) return null;
		if (root.classList && root.classList.contains('duration-bridge')) {
			return root;
		}
		return root.querySelector('.duration-bridge');
	}

	function setDurationBridgeState(bridge, label, invalid) {
		if (!bridge) return;
		var pill = bridge.querySelector('.duration-pill');
		var valueEl = bridge.querySelector('.duration-pill-value');
		var icon = bridge.querySelector('.duration-pill-icon');
		bridge.removeAttribute('hidden');
		if (!label) {
			bridge.classList.remove('duration-bridge--invalid');
			bridge.setAttribute('aria-label', 'Duration');
			if (pill) pill.className = 'pill pill--neutral duration-pill duration-pill--pending';
			if (icon) icon.hidden = false;
			if (valueEl) valueEl.textContent = '-';
			return;
		}
		if (invalid) {
			bridge.classList.add('duration-bridge--invalid');
			bridge.setAttribute('aria-label', label);
		} else {
			bridge.classList.remove('duration-bridge--invalid');
			bridge.setAttribute('aria-label', 'Duration: ' + label);
		}
		if (pill) {
			pill.className = 'pill duration-pill ' + (invalid ? 'pill--error' : 'pill--neutral');
		}
		if (icon) {
			icon.hidden = !!invalid;
		}
		if (valueEl) {
			valueEl.textContent = invalid ? 'Invalid' : label;
		}
	}

	function refreshTimeRangePreview(opts) {
		var targets = resolveTimeRangeTargets(opts);
		var startInput = targets.startInput;
		var endInput = targets.endInput;
		var preview = targets.preview;
		if (!startInput || !endInput || !preview) return;

		var start = parseTimeToMinutes(startInput.value);
		var end = parseTimeToMinutes(endInput.value);
		if (start === null || end === null || !startInput.value || !endInput.value) {
			setDurationBridgeState(preview, '', false);
			return;
		}
		if (end <= start) {
			setDurationBridgeState(preview, 'End time must be after start time', true);
			return;
		}
		setDurationBridgeState(preview, formatDurationLabel(end - start), false);
	}

	function validateTimeRange(opts) {
		opts = opts || {};
		var targets = resolveTimeRangeTargets(opts);
		var startInput = targets.startInput;
		var endInput = targets.endInput;
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

	function showTimeRangeError(message, previewEl) {
		var bridge = findDurationBridge(previewEl);
		if (!bridge && previewEl) {
			bridge = previewEl;
		}
		if (!bridge) {
			bridge = document.getElementById('durationPreview');
		}
		setDurationBridgeState(bridge, message, true);
	}

	function initTimeRangePreview(opts) {
		opts = opts || {};
		var targets = resolveTimeRangeTargets(opts);
		var startInput = targets.startInput;
		var endInput = targets.endInput;
		var preview = targets.preview;
		var form = opts.form || (startInput ? startInput.closest('form') : null);

		if (form && form.dataset.timeRangePreview === 'true') {
			refreshTimeRangePreview({ startInput: startInput, endInput: endInput, preview: preview });
			return;
		}
		if (form) {
			form.dataset.timeRangePreview = 'true';
		}
		if (!startInput || !endInput || !preview) return;

		function update() {
			refreshTimeRangePreview({ startInput: startInput, endInput: endInput, preview: preview });
		}

		startInput.addEventListener('input', update);
		startInput.addEventListener('change', update);
		endInput.addEventListener('input', update);
		endInput.addEventListener('change', update);

		if (form) {
			form.addEventListener('reset', function () {
				window.requestAnimationFrame(update);
			});
		}

		update();
	}

	function initAllTimeRangePreviews() {
		document.querySelectorAll('form').forEach(function (form) {
			if (!form.querySelector('[name="start_time"]') || !form.querySelector('[name="end_time"]') || !form.querySelector('.duration-bridge')) {
				return;
			}
			initTimeRangePreview({ form: form });
		});
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
				var startInput = form.querySelector('[name="start_time"]');
				var endInput = form.querySelector('[name="end_time"]');
				var preview = form.querySelector('.duration-bridge');
				var result = validateTimeRange({ startInput: startInput, endInput: endInput });
				if (!result.ok) {
					e.preventDefault();
					e.stopImmediatePropagation();
					showTimeRangeError(result.message, preview);
					alert(result.message);
				}
			}, true);
		});
	}

	window.validateTimeRange = validateTimeRange;
	window.refreshTimeRangePreview = refreshTimeRangePreview;
	window.initTimeRangePreview = initTimeRangePreview;
	window.initAllTimeRangePreviews = initAllTimeRangePreviews;
	window.attachTimeRangeFormValidation = attachTimeRangeFormValidation;
})();
