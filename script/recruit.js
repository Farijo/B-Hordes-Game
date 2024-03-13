(function() {
    const startDateInput = $('[name=start_date]');
    startDateInput.on('change', function () {
        if(!this.value || new Date(this.value) < new Date()) {
            startDateInput.siblings('input').val('Démarrer maintenant');
            startDateInput.parent().attr('onsubmit', "return confirm('Le lancement du défi empechera toute modification, hormis la date de fin')");
        } else {
            startDateInput.siblings('input').val('Valider');
            startDateInput.parent().attr('onsubmit', "dateToISOGMT(this)");
        }
    }).trigger('change');

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
