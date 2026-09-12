(function () {
	'use strict';

	var instances = {};

	function cssVar(name) {
		return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
	}

	function chartColors() {
		return {
			success: cssVar('--color-success') || '#059669',
			destructive: cssVar('--color-destructive') || '#DC2626',
			warning: cssVar('--color-warning') || '#D97706',
			muted: cssVar('--color-muted-foreground') || '#64748B',
			border: cssVar('--color-border') || '#E2E8F0',
			foreground: cssVar('--color-foreground') || '#0F172A',
			mutedForeground: cssVar('--color-muted-foreground') || '#64748B'
		};
	}

	function destroyChart(key) {
		if (instances[key]) {
			instances[key].destroy();
			delete instances[key];
		}
	}

	function setEmptyState(wrap, message) {
		if (!wrap) return;
		var canvas = wrap.querySelector('canvas');
		if (canvas) canvas.hidden = true;
		var existing = wrap.querySelector('.analytics-chart-empty');
		if (existing) {
			existing.textContent = message;
			return;
		}
		var empty = document.createElement('p');
		empty.className = 'analytics-chart-empty';
		empty.textContent = message;
		wrap.appendChild(empty);
	}

	function clearEmptyState(wrap) {
		if (!wrap) return;
		var canvas = wrap.querySelector('canvas');
		if (canvas) canvas.hidden = false;
		var existing = wrap.querySelector('.analytics-chart-empty');
		if (existing) existing.remove();
	}

	function baseOptions(colors) {
		return {
			responsive: true,
			maintainAspectRatio: false,
			plugins: {
				legend: {
					position: 'bottom',
					labels: {
						color: colors.foreground,
						boxWidth: 12,
						padding: 16
					}
				},
				tooltip: {
					callbacks: {
						label: function (context) {
							var label = context.dataset.label || context.label || '';
							if (label) label += ': ';
							var value = context.parsed.y != null ? context.parsed.y : context.parsed.x;
							if (value == null && context.parsed != null) value = context.parsed;
							label += value;
							return label;
						}
					}
				}
			}
		};
	}

	function renderClassStatus(summary) {
		var canvas = document.getElementById('classStatusChart');
		var wrap = canvas && canvas.closest('.analytics-chart-wrap');
		if (!canvas || !wrap) return;

		destroyChart('classStatus');
		var conducted = Number(summary && summary.conducted) || 0;
		var cancelled = Number(summary && summary.cancelled) || 0;
		var rescheduled = Number(summary && summary.rescheduled) || 0;

		if (conducted + cancelled + rescheduled === 0) {
			setEmptyState(wrap, 'No class records in this period.');
			return;
		}

		clearEmptyState(wrap);
		var colors = chartColors();
		instances.classStatus = new Chart(canvas, {
			type: 'doughnut',
			data: {
				labels: ['Conducted', 'Cancelled', 'Rescheduled'],
				datasets: [{
					data: [conducted, cancelled, rescheduled],
					backgroundColor: [colors.success, colors.destructive, colors.warning],
					borderColor: colors.border,
					borderWidth: 1
				}]
			},
			options: Object.assign({}, baseOptions(colors), {
				plugins: Object.assign({}, baseOptions(colors).plugins, {
					tooltip: {
						callbacks: {
							label: function (context) {
								var total = context.dataset.data.reduce(function (sum, value) {
									return sum + value;
								}, 0);
								var value = context.parsed;
								var pct = total > 0 ? ((value / total) * 100).toFixed(1) : '0.0';
								return context.label + ': ' + value + ' (' + pct + '%)';
							}
						}
					}
				})
			})
		});
	}

	function renderWeeklyTrend(weekly) {
		var canvas = document.getElementById('weeklyTrendChart');
		var wrap = canvas && canvas.closest('.analytics-chart-wrap');
		if (!canvas || !wrap) return;

		destroyChart('weeklyTrend');
		var rows = Array.isArray(weekly) ? weekly.slice() : [];
		rows.sort(function (left, right) {
			return String(left.weekLabel || '').localeCompare(String(right.weekLabel || ''));
		});

		if (rows.length === 0) {
			setEmptyState(wrap, 'No weekly data in this period.');
			return;
		}

		clearEmptyState(wrap);
		var colors = chartColors();
		var labels = rows.map(function (row) { return row.weekLabel || ''; });
		instances.weeklyTrend = new Chart(canvas, {
			type: 'bar',
			data: {
				labels: labels,
				datasets: [
					{
						label: 'Conducted',
						data: rows.map(function (row) { return row.conducted || 0; }),
						backgroundColor: colors.success,
						borderRadius: 4
					},
					{
						label: 'Cancelled',
						data: rows.map(function (row) { return row.cancelled || 0; }),
						backgroundColor: colors.destructive,
						borderRadius: 4
					},
					{
						label: 'Rescheduled',
						data: rows.map(function (row) { return row.rescheduled || 0; }),
						backgroundColor: colors.warning,
						borderRadius: 4
					}
				]
			},
			options: Object.assign({}, baseOptions(colors), {
				scales: {
					x: {
						ticks: { color: colors.mutedForeground },
						grid: { color: colors.border }
					},
					y: {
						beginAtZero: true,
						ticks: {
							color: colors.mutedForeground,
							precision: 0
						},
						grid: { color: colors.border }
					}
				}
			})
		});
	}

	function topInactiveReasons(rows) {
		var sorted = rows.slice().sort(function (left, right) {
			return (right.count || 0) - (left.count || 0);
		});
		if (sorted.length <= 8) return sorted;
		var top = sorted.slice(0, 7);
		var otherCount = sorted.slice(7).reduce(function (sum, row) {
			return sum + (row.count || 0);
		}, 0);
		if (otherCount > 0) {
			top.push({ reason: 'Other', count: otherCount });
		}
		return top;
	}

	function renderInactiveReasons(rows) {
		var canvas = document.getElementById('inactiveReasonsChart');
		var wrap = canvas && canvas.closest('.analytics-chart-wrap');
		if (!canvas || !wrap) return;

		destroyChart('inactiveReasons');
		var reasons = Array.isArray(rows) ? topInactiveReasons(rows) : [];
		if (reasons.length === 0) {
			setEmptyState(wrap, 'No inactive students.');
			return;
		}

		clearEmptyState(wrap);
		var colors = chartColors();
		instances.inactiveReasons = new Chart(canvas, {
			type: 'bar',
			data: {
				labels: reasons.map(function (row) { return row.reason || '(not specified)'; }),
				datasets: [{
					label: 'Students',
					data: reasons.map(function (row) { return row.count || 0; }),
					backgroundColor: colors.success,
					borderRadius: 4
				}]
			},
			options: Object.assign({}, baseOptions(colors), {
				indexAxis: 'y',
				plugins: Object.assign({}, baseOptions(colors).plugins, {
					legend: { display: false }
				}),
				scales: {
					x: {
						beginAtZero: true,
						ticks: {
							color: colors.mutedForeground,
							precision: 0
						},
						grid: { color: colors.border }
					},
					y: {
						ticks: { color: colors.mutedForeground },
						grid: { display: false }
					}
				}
			})
		});
	}

	window.analyticsCharts = {
		destroyAll: function () {
			Object.keys(instances).forEach(function (key) {
				destroyChart(key);
			});
		},
		renderAll: function (data) {
			if (!window.Chart) return;
			renderClassStatus(data && data.summary);
			renderWeeklyTrend(data && data.weekly);
			renderInactiveReasons(data && data.inactiveReasons);
		}
	};
})();
