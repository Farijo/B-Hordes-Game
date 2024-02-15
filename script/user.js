(function () {
    const table = $('.sortable');
    const rows = table.find('.item');
    const select = $('#type');
    select.on('change', function(i, e) {
        const val = this.value;
        rows.each(function() {
            const r = $(this);
            if(r.children(':first').attr('title').includes(val)) {
                table.append(r);
            } else {
                r.remove();
            }
        });
    }).val(new URLSearchParams(window.location.search).get('selection') || select.children(':first').val()).trigger('change');
})();
