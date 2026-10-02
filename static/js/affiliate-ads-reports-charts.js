(function () {
	'use strict';

	var instances = {};
	var BAR_ROW_HEIGHT = 32;
	var BAR_CHART_PADDING = 56;
	var BAR_CHART_MIN_HEIGHT = 140;
	var BAR_CHART_MAX_HEIGHT = 380;
	var MAX_BAR_THICKNESS = 22;

	function cssVar(name) {
		return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
	}

	function chartColors() {
		return {
			primary: cssVar('--color-primary') || '#90C020',
			primaryLight: cssVar('--color-primary-light') || '#DCFCE7',
			info: cssVar('--color-info') || '#0EA5E9',
			infoBg: cssVar('--color-info-bg') || '#E0F2FE',
			success: cssVar('--color-success') || '#059669',
			successBg: cssVar('--color-success-bg') || '#D1FAE5',
			muted: cssVar('--color-muted-foreground') || '#64748B',
			border: cssVar('--color-border') || '#E2E8F0',
			foreground: cssVar('--color-foreground') || '#0F172A',
			mutedForeground: cssVar('--color-muted-foreground') || '#64748B'
		};
	}

	function hexToRgba(hex, alpha) {
		var h = (hex || '').replace('#', '');
		if (h.length === 3) {
			h = h[0] + h[0] + h[1] + h[1] + h[2] + h[2];
		}
		if (h.length !== 6) {
			return 'rgba(144, 192, 32, ' + alpha + ')';
		}
		var r = parseInt(h.slice(0, 2), 16);
		var g = parseInt(h.slice(2, 4), 16);
		var b = parseInt(h.slice(4, 6), 16);
		return 'rgba(' + r + ', ' + g + ', ' + b + ', ' + alpha + ')';
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
		var existing = wrap.querySelector('.affiliate-ads-reports-chart-empty');
		if (existing) {
			existing.textContent = message;
			return;
		}
		var empty = document.createElement('p');
		empty.className = 'affiliate-ads-reports-chart-empty';
		empty.textContent = message;
		wrap.appendChild(empty);
	}

	function clearEmptyState(wrap) {
		if (!wrap) return;
		var canvas = wrap.querySelector('canvas');
		if (canvas) canvas.hidden = false;
		var existing = wrap.querySelector('.affiliate-ads-reports-chart-empty');
		if (existing) existing.remove();
	}

	function setBarChartHeight(wrap, itemCount) {
		if (!wrap) return;
		var count = Math.max(1, itemCount);
		var height = Math.min(
			BAR_CHART_MAX_HEIGHT,
			Math.max(BAR_CHART_MIN_HEIGHT, count * BAR_ROW_HEIGHT + BAR_CHART_PADDING)
		);
		wrap.style.height = height + 'px';
	}

	function truncateTickLabel(label, maxLen) {
		var text = String(label || '');
		if (text.length <= maxLen) return text;
		if (maxLen <= 3) return text.slice(0, maxLen);
		return text.slice(0, maxLen - 3) + '...';
	}

	function baseOptions(colors) {
		return {
			responsive: true,
			maintainAspectRatio: false,
			layout: {
				padding: { top: 4, right: 8, bottom: 4, left: 4 }
			},
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

	function barFillColor(colorKey, colors) {
		switch (colorKey) {
			case 'info':
				return colors.info;
			case 'success':
				return colors.success;
			default:
				return colors.primary;
		}
	}

	function renderHorizontalBar(canvasId, rows, emptyMessage, colorKey) {
		var canvas = document.getElementById(canvasId);
		var wrap = canvas && canvas.closest('.affiliate-ads-reports-chart-wrap');
		if (!canvas || !wrap) return;

		destroyChart(canvasId);
		var items = Array.isArray(rows) ? rows : [];
		if (items.length === 0) {
			wrap.style.height = BAR_CHART_MIN_HEIGHT + 'px';
			setEmptyState(wrap, emptyMessage);
			return;
		}

		clearEmptyState(wrap);
		setBarChartHeight(wrap, items.length);

		var colors = chartColors();
		var barColor = barFillColor(colorKey, colors);
		var labels = items.map(function (row) { return row.label || ''; });

		instances[canvasId] = new Chart(canvas, {
			type: 'bar',
			data: {
				labels: labels,
				datasets: [{
					label: 'Clicks',
					data: items.map(function (row) { return row.clicks || 0; }),
					backgroundColor: hexToRgba(barColor, 0.88),
					hoverBackgroundColor: barColor,
					borderColor: barColor,
					borderWidth: 1,
					borderRadius: 6,
					borderSkipped: false,
					maxBarThickness: MAX_BAR_THICKNESS,
					barThickness: MAX_BAR_THICKNESS
				}]
			},
			options: Object.assign({}, baseOptions(colors), {
				indexAxis: 'y',
				plugins: Object.assign({}, baseOptions(colors).plugins, {
					legend: { display: false },
					tooltip: {
						callbacks: {
							title: function (context) {
								if (!context.length) return '';
								return labels[context[0].dataIndex] || '';
							},
							label: function (context) {
								return 'Clicks: ' + context.parsed.x;
							}
						}
					}
				}),
				scales: {
					x: {
						beginAtZero: true,
						ticks: {
							color: colors.mutedForeground,
							precision: 0,
							font: { size: 11 }
						},
						grid: { color: colors.border, drawBorder: false }
					},
					y: {
						ticks: {
							color: colors.mutedForeground,
							font: { size: 11 },
							autoSkip: false,
							callback: function (value) {
								var label = this.getLabelForValue(value);
								return truncateTickLabel(label, 48);
							}
						},
						grid: { display: false },
						border: { display: false }
					}
				}
			})
		});
	}

	function renderProductActivity(summary) {
		var canvas = document.getElementById('affiliateProductActivityChart');
		var wrap = canvas && canvas.closest('.affiliate-ads-reports-chart-wrap');
		if (!canvas || !wrap) return;

		destroyChart('affiliateProductActivityChart');
		var withClicks = Number(summary && summary.productsWithClicks) || 0;
		var withoutClicks = Number(summary && summary.productsWithoutClicks) || 0;
		if (withClicks + withoutClicks === 0) {
			wrap.style.height = '220px';
			setEmptyState(wrap, 'No affiliate products yet.');
			return;
		}

		clearEmptyState(wrap);
		wrap.style.height = '220px';
		var colors = chartColors();
		instances.affiliateProductActivityChart = new Chart(canvas, {
			type: 'doughnut',
			data: {
				labels: ['With clicks', 'No clicks yet'],
				datasets: [{
					data: [withClicks, withoutClicks],
					backgroundColor: [colors.success, colors.border],
					borderColor: colors.border,
					borderWidth: 1,
					hoverOffset: 4
				}]
			},
			options: Object.assign({}, baseOptions(colors), {
				cutout: '62%',
				plugins: Object.assign({}, baseOptions(colors).plugins, {
					legend: {
						position: 'bottom',
						labels: {
							color: colors.foreground,
							boxWidth: 10,
							padding: 12,
							font: { size: 11 }
						}
					},
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

	window.affiliateAdsReportsCharts = {
		renderAll: function (data) {
			if (!window.Chart || !data) return;
			renderHorizontalBar(
				'topAffiliateProductsChart',
				data.topProducts,
				'No affiliate link clicks recorded yet.',
				'primary'
			);
			renderHorizontalBar(
				'topAdsChart',
				data.topAds,
				'No linked ad click totals yet. Clicks count at the product level.',
				'info'
			);
			renderHorizontalBar(
				'topTeachersChart',
				data.topTeachers,
				'No teacher click events recorded yet.',
				'success'
			);
			renderHorizontalBar(
				'clicksByAdZoneChart',
				data.clicksByAdZone,
				'No attributed ad zone clicks yet.',
				'primary'
			);
			renderHorizontalBar(
				'clicksByProviderChart',
				data.clicksByProvider,
				'No provider click totals yet.',
				'info'
			);
			renderProductActivity(data.summary);
		}
	};
})();
