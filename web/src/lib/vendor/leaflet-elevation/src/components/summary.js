const _ = L.Control.Elevation.Utils;

export var Summary = L.Class.extend({
	initialize(opts, control) {
		this.options = opts;
		this.control = control;
		this.labels = {};

		let summary = this._container = _.create("div", "elevation-summary " + (opts.summary ? opts.summary + "-summary" : ''));
		_.style(summary, 'max-width', opts.width ? opts.width + 'px' : '');
	},

	render() {
		return container => container.append(() => this._container);
	},

	reset() {
		this._container.innerHTML = '';
	},

	append(className, label, value) {
		this._container.innerHTML += `<span class="${className}"><span class="summarylabel">${label}</span><span class="summaryvalue">${value}</span></span>`;
		return this;
	},

	update() {
		Object
		.keys(this.labels)
		.sort((a, b) => this.labels[a].order - this.labels[b].order) // TODO: any performance issues?
		.forEach((i)=> {
			let value = typeof this.labels[i].value !== "function" ? this.labels[i].value : this.labels[i].value(this.control.track_info, this.labels[i].unit || '');
			this.append(i /*+ " order-" + this.labels[i].order*/, L._(this.labels[i].label), value, this.labels[i].order);
		});
	},

	_registerSummary(data) {
		for (let i in data) {
			data[i].order = data[i].order ?? 1000;
			this.labels[i] = data[i];
		}
	}

});
