(function () {
	const container = document.getElementById('student-relationship-graph');
	const emptyState = document.getElementById('student-relationships-empty');
	const searchInput = document.getElementById('diagramStudentSearch');
	const searchResults = document.getElementById('diagramStudentSearchResults');
	const searchFeedback = document.getElementById('student-relationships-search-feedback');
	const controlsOverlay = document.querySelector('.student-relationships-controls-overlay');
	const legendOverlay = document.querySelector('.student-relationships-legend-overlay');
	const prevButton = document.getElementById('diagramStudentPrev');
	const nextButton = document.getElementById('diagramStudentNext');
	const focusValue = document.getElementById('diagramStudentFocusValue');
	if (!container) return;

	const apiURL = container.getAttribute('data-api-url');
	if (!apiURL || typeof ForceGraph !== 'function') {
		container.innerHTML = '<p class="student-relationships-loading">Unable to load diagram.</p>';
		return;
	}

	const colors = {
		text: '#0f172a',
		border: '#cbd5e1',
		parentFill: '#f1f5f9',
		parentBorder: '#94a3b8',
		link: '#64748b',
		labelBg: '#ffffff',
		labelText: '#475569',
		focus: '#90C020',
		inactiveOpacity: 0.55,
	};

	let graphInstance = null;
	let focusedNodeId = null;
	let graphNodes = [];
	let studentNodes = [];
	let focusedStudentIndex = -1;

	function setSearchFeedback(message, isError) {
		if (!searchFeedback) return;
		if (!isError || !message) {
			searchFeedback.textContent = '';
			return;
		}
		searchFeedback.textContent = message;
	}

	function refreshGraph() {
		if (graphInstance && typeof graphInstance.refresh === 'function') {
			graphInstance.refresh();
		}
	}

	function rebuildStudentNodes() {
		studentNodes = graphNodes
			.filter(function (node) {
				return node.kind === 'student';
			})
			.sort(function (a, b) {
				return (a.name || '').localeCompare(b.name || '');
			});
	}

	function updateNavControls() {
		const count = studentNodes.length;
		if (prevButton) {
			prevButton.disabled = count === 0;
		}
		if (nextButton) {
			nextButton.disabled = count === 0;
		}
		if (!focusValue) {
			return;
		}
		if (count === 0) {
			focusValue.textContent = 'No students in diagram';
			return;
		}
		if (focusedStudentIndex < 0) {
			focusValue.textContent = 'No student focused';
			return;
		}
		const node = studentNodes[focusedStudentIndex];
		focusValue.textContent = node.name + ' (' + (focusedStudentIndex + 1) + ' of ' + count + ')';
	}

	function centerOnNode(node) {
		const focusNow = function () {
			if (!graphInstance || !node) {
				return;
			}
			if (node.x == null || node.y == null) {
				window.setTimeout(focusNow, 50);
				return;
			}
			graphInstance.centerAt(node.x, node.y, 700);
			graphInstance.zoom(2.2, 700);
		};
		focusNow();
	}

	function applyFocusToNode(node) {
		if (!node) {
			return;
		}
		focusedNodeId = node.id;
		focusedStudentIndex = studentNodes.findIndex(function (item) {
			return item.id === node.id;
		});
		updateNavControls();
		setSearchFeedback('', false);
		refreshGraph();
		if (searchInput) {
			searchInput.value = node.name;
		}
		if (searchResults) {
			searchResults.innerHTML = '';
		}
		centerOnNode(node);
	}

	function focusStudentAtIndex(index) {
		if (!studentNodes.length) {
			return;
		}
		const normalized = ((index % studentNodes.length) + studentNodes.length) % studentNodes.length;
		applyFocusToNode(studentNodes[normalized]);
	}

	function focusStudentNode(studentId, studentName) {
		if (!graphInstance) {
			setSearchFeedback('Diagram is still loading. Try again in a moment.', true);
			return;
		}

		const nodeId = String(studentId);
		const node = studentNodes.find(function (item) {
			return item.id === nodeId;
		});

		if (!node) {
			setSearchFeedback(
				(studentName || 'Student') + ' is not shown in the diagram. Add a parent name or relationship on their profile.',
				true
			);
			return;
		}

		applyFocusToNode(node);
	}

	function goToPreviousStudent() {
		if (!studentNodes.length) {
			return;
		}
		if (focusedStudentIndex < 0) {
			focusStudentAtIndex(studentNodes.length - 1);
			return;
		}
		focusStudentAtIndex(focusedStudentIndex - 1);
	}

	function goToNextStudent() {
		if (!studentNodes.length) {
			return;
		}
		if (focusedStudentIndex < 0) {
			focusStudentAtIndex(0);
			return;
		}
		focusStudentAtIndex(focusedStudentIndex + 1);
	}

	function openStudentViewModal(node) {
		if (!node || node.kind !== 'student' || !node.viewUrl) {
			return;
		}
		const host = document.getElementById('studentViewModalHost');
		if (!host) {
			return;
		}
		applyFocusToNode(node);
		if (window.htmx) {
			window.htmx.ajax('GET', node.viewUrl, {
				target: '#studentViewModalHost',
				swap: 'innerHTML',
			});
			return;
		}
		fetch(node.viewUrl, { credentials: 'same-origin' })
			.then(function (response) {
				if (!response.ok) {
					throw new Error('Failed to load student view');
				}
				return response.text();
			})
			.then(function (html) {
				host.innerHTML = html;
				document.body.classList.add('modal-open');
				const closeBtn = host.querySelector('.student-view-close');
				if (closeBtn) {
					closeBtn.focus();
				}
			})
			.catch(function () {
				setSearchFeedback('Failed to open student details.', true);
			});
	}

	function measureNode(ctx, node, globalScale) {
		const fontSize = (node.kind === 'parent' ? 11 : 12) / globalScale;
		const paddingX = 10 / globalScale;
		const paddingY = 7 / globalScale;
		ctx.font = `${node.kind === 'parent' ? '600 ' : ''}${fontSize}px "Open Sans", sans-serif`;
		const textWidth = ctx.measureText(node.name || '').width;
		node.__boxW = textWidth + paddingX * 2;
		node.__boxH = fontSize + paddingY * 2;
	}

	function drawRoundedRect(ctx, x, y, width, height, radius) {
		const r = Math.min(radius, width / 2, height / 2);
		ctx.beginPath();
		ctx.moveTo(x + r, y);
		ctx.lineTo(x + width - r, y);
		ctx.quadraticCurveTo(x + width, y, x + width, y + r);
		ctx.lineTo(x + width, y + height - r);
		ctx.quadraticCurveTo(x + width, y + height, x + width - r, y + height);
		ctx.lineTo(x + r, y + height);
		ctx.quadraticCurveTo(x, y + height, x, y + height - r);
		ctx.lineTo(x, y + r);
		ctx.quadraticCurveTo(x, y, x + r, y);
		ctx.closePath();
	}

	function drawNodeBox(node, ctx, globalScale) {
		measureNode(ctx, node, globalScale);
		const width = node.__boxW;
		const height = node.__boxH;
		const x = node.x - width / 2;
		const y = node.y - height / 2;
		const radius = 6 / globalScale;
		const isParent = node.kind === 'parent';
		const isInactive = node.kind === 'student' && node.status === 'inactive';
		const isFocused = node.id === focusedNodeId;

		ctx.save();
		if (isInactive) {
			ctx.globalAlpha = colors.inactiveOpacity;
		}

		if (isFocused) {
			ctx.shadowColor = colors.focus;
			ctx.shadowBlur = 14 / globalScale;
		}

		drawRoundedRect(ctx, x, y, width, height, radius);
		ctx.fillStyle = isParent ? colors.parentFill : '#ffffff';
		ctx.fill();

		ctx.shadowBlur = 0;
		ctx.lineWidth = (isFocused ? 3 : isParent ? 1.5 : 2) / globalScale;
		if (isParent) {
			ctx.setLineDash([4 / globalScale, 3 / globalScale]);
			ctx.strokeStyle = isFocused ? colors.focus : colors.parentBorder;
		} else {
			ctx.setLineDash([]);
			ctx.strokeStyle = isFocused ? colors.focus : (node.color || colors.border);
		}
		ctx.stroke();
		ctx.setLineDash([]);

		ctx.fillStyle = colors.text;
		ctx.textAlign = 'center';
		ctx.textBaseline = 'middle';
		const fontSize = (isParent ? 11 : 12) / globalScale;
		ctx.font = `${isParent ? '600 ' : ''}${fontSize}px "Open Sans", sans-serif`;
		ctx.fillText(node.name || '', node.x, node.y);
		ctx.restore();
	}

	function drawLinkLabel(link, ctx, globalScale) {
		if (!link.label) return;

		const source = link.source;
		const target = link.target;
		if (!source || !target || source.x == null || target.x == null) return;

		const midX = (source.x + target.x) / 2;
		const midY = (source.y + target.y) / 2;
		const fontSize = 10 / globalScale;
		const paddingX = 5 / globalScale;
		const paddingY = 3 / globalScale;

		ctx.font = `${fontSize}px "Open Sans", sans-serif`;
		const textWidth = ctx.measureText(link.label).width;
		const boxW = textWidth + paddingX * 2;
		const boxH = fontSize + paddingY * 2;

		ctx.fillStyle = colors.labelBg;
		ctx.strokeStyle = colors.border;
		ctx.lineWidth = 1 / globalScale;
		drawRoundedRect(ctx, midX - boxW / 2, midY - boxH / 2, boxW, boxH, 4 / globalScale);
		ctx.fill();
		ctx.stroke();

		ctx.fillStyle = colors.labelText;
		ctx.textAlign = 'center';
		ctx.textBaseline = 'middle';
		ctx.fillText(link.label, midX, midY);
	}

	if (searchResults) {
		searchResults.addEventListener('click', function (event) {
			const item = event.target.closest('.entity-search-item');
			if (!item) return;
			event.preventDefault();
			focusStudentNode(item.getAttribute('data-id'), item.getAttribute('data-name'));
		});
	}

	if (searchInput) {
		searchInput.addEventListener('keydown', function (event) {
			if (event.key !== 'Enter') return;
			event.preventDefault();
			const firstItem = searchResults && searchResults.querySelector('.entity-search-item');
			if (firstItem) {
				focusStudentNode(firstItem.getAttribute('data-id'), firstItem.getAttribute('data-name'));
			}
		});
	}

	if (prevButton) {
		prevButton.addEventListener('click', goToPreviousStudent);
	}

	if (nextButton) {
		nextButton.addEventListener('click', goToNextStudent);
	}

	fetch(apiURL, { credentials: 'same-origin' })
		.then(function (response) {
			if (!response.ok) {
				throw new Error('Failed to load graph data');
			}
			return response.json();
		})
		.then(function (data) {
			const nodes = data.nodes || [];
			const links = data.links || [];
			graphNodes = nodes;
			rebuildStudentNodes();
			updateNavControls();
			container.innerHTML = '';

			if (!nodes.length) {
				if (emptyState) {
					emptyState.classList.add('is-visible');
				}
				if (controlsOverlay) {
					controlsOverlay.hidden = true;
				}
				if (legendOverlay) {
					legendOverlay.hidden = true;
				}
				return;
			}

			if (emptyState) {
				emptyState.classList.remove('is-visible');
			}
			if (controlsOverlay) {
				controlsOverlay.hidden = false;
			}
			if (legendOverlay) {
				legendOverlay.hidden = false;
			}

			const graph = ForceGraph()(container)
				.graphData({ nodes: nodes, links: links })
				.nodeId('id')
				.linkSource('source')
				.linkTarget('target')
				.backgroundColor('rgba(0,0,0,0)')
				.linkColor(function () { return colors.link; })
				.linkWidth(1.5)
				.linkDirectionalArrowLength(5)
				.linkDirectionalArrowRelPos(0.92)
				.linkCurvature(0.12)
				.linkCanvasObjectMode(function () { return 'after'; })
				.linkCanvasObject(drawLinkLabel)
				.nodeCanvasObjectMode(function () { return 'replace'; })
				.nodeCanvasObject(drawNodeBox)
				.onNodeClick(function (node) {
					openStudentViewModal(node);
				})
				.cooldownTicks(120)
				.d3AlphaDecay(0.03)
				.d3VelocityDecay(0.35);

			graph.d3Force('charge').strength(function (node) {
				return node.kind === 'parent' ? -420 : -280;
			});
			graph.d3Force('link').distance(function (link) {
				return link.label === 'child' ? 90 : 130;
			});

			graphInstance = graph;
		})
		.catch(function () {
			container.innerHTML = '<p class="student-relationships-loading">Failed to load diagram.</p>';
			if (controlsOverlay) {
				controlsOverlay.hidden = true;
			}
			if (legendOverlay) {
				legendOverlay.hidden = true;
			}
		});
})();
