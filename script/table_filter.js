(function () {
    $('.sortable').each(function(i, e) {
        const t = $(e);
        const tr = t.find('.item');
        const selects = t.find('select');
        selects.on('change', function () {

            const values = [];
            selects.each(function() {
                if(this.value != '*') {
                    values.push(this.value);
                }
            });

            if (values.length) {
                tr.each(function() {
                    const r = $(this);
                    if(values.every(value => r.find(`[title="${value}"]`).length > 0)) {
                        r.show();
                    } else {
                        r.hide();
                    }
                });
            } else {
                tr.show();
            }
        });
    })

    $('td.date').each((i, e) => {
        const date = new Date(e.getAttribute("sorttable_customkey")+'Z');
        e.innerHTML = `${date.toLocaleDateString()}<br><span style=\"font-size:80%\">${date.toLocaleTimeString()}</span>`;
    })
})();
