(function () {
	'use strict';

	var timeOptions = { hour: '2-digit', minute: '2-digit', second: '2-digit' };

	function updateClocks() {
		var now = new Date();
		var clocks = document.querySelectorAll('.digital-clock');
		for (var i = 0; i < clocks.length; i++) {
			var clock = clocks[i];
			clock.dateTime = now.toISOString();
			clock.textContent = now.toLocaleTimeString(undefined, timeOptions);
		}
	}

	function init() {
		updateClocks();
		setInterval(updateClocks, 1000);
	}

	if (document.readyState === 'loading') {
		document.addEventListener('DOMContentLoaded', init);
	} else {
		init();
	}
})();
