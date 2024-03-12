(function() {
    const isoNow = new Date().toLocaleString('sv').substring(0, 16);
    
    const sdv = $('[name=valider]');
    $('[name=start_date]').on('change', function () {
        if(!this.value || new Date(this.value) < new Date()) {
            sdv.val('Démarrer maintenant');
            sdv.parent().attr('onsubmit', "return confirm('Le lancement du défi empechera toute modification')");
        } else {
            sdv.val('Valider');
            sdv.parent().attr('onsubmit', "dateToISOGMT(this)");
        }
    }).trigger('change');

    $('[name=end_date]').prop('min', isoNow);

    $('[type=checkbox]').on('change', function () {
        if(this.checked) {
            $(`#${this.attributes.form.textContent} [type=submit]`).prop('disabled', false);
        } else if($('[form="Résultats"]:checked').length == 0) {
            $(`#${this.attributes.form.textContent} [type=submit]`).prop('disabled', true);
        }
    });

    function decomposeTemps(duree) {
        return '~' + String(Math.floor(duree / 3600)).padStart(2, '0') + ':' +
               String(Math.floor((duree % 3600) / 60)).padStart(2, '0') + ':' +
               String(duree % 60).padStart(2, '0');
    }

    const it = [];
    $('[data-rem]').each((i, e) => {
        let rem = e.dataset.rem;
        console.time('r');
        it[i] = setInterval(() => {
            rem--;
            if(rem <= 0) {
                clearInterval(it[i]);
                window.location.reload();
            }
            e.innerText = decomposeTemps(rem);
        }, 1000);
    });
})();

function childClick(event) {
    event.stopPropagation();
    // const c=event.target.children[0];
    // if(c) {
        // $(c).click();
    // }
}

function dateToISOGMT(e) {
    const input = $(e).children('[type="datetime-local"]');
    input.val(new Date(input.val()).toISOString().substring(0, 16));
}
