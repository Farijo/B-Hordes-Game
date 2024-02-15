(function () {

    const info = $('.info');
    info.hide();

    $('.toggle-btn').on('click', function() {
        var $this = $(this);
        var subJSON = $this.next('.sub-json');
        subJSON.toggle();
        $this.text(subJSON.is(':visible') ? 'Masquer' : 'Afficher');
    });

    $('tr:not(.info)').on('click', function () {
        $(this).next('tr').toggle();
    });
})();
