(function() {
    const startDateInput = $('[name=start_date]');

    const endDateInput = $('[name=end_date]');
    if(typeof startDate === "string") {
        startDateInput.val(startDate);
        endDateInput.prop('min', startDate);
        const sdd = new Date(startDate);
        const p = $('p').first();
        p.text(sdd.toLocaleString());
        let rem = Math.round((sdd.getTime() - Date.now())/1000);
        if(rem < 24*60*60) {
            const sp = p.siblings('span');
            const it = setInterval(() => {
                rem--;
                if(rem <= 0) {
                    clearInterval(it);
                    window.location.reload();
                }
                sp.text(decomposeTemps(rem));
            }, 1000);
        }
    } else {
        endDateInput.parent().prop('title', "Une date de début doit d'abord être renseignée").children('input').prop('disabled', true)
    }
    if(typeof endDate === "string") {
        startDateInput.prop('max', endDate);
        endDateInput.val(endDate);
        const edd = new Date(endDate);
        const p = $('p').last();
        p.text(edd.toLocaleString());
        let rem = edd.getTime() - Date.now();
        if(rem < 24*60*60*1000) {
            const sp = p.siblings('span');
            const it = setInterval(() => {
                rem--;
                if(rem <= 0) {
                    clearInterval(it);
                    window.location.reload();
                }
                sp.text(decomposeTemps(rem));
            }, 1000);
        }
    }

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
        return '~ ' + String(Math.floor(duree / 3600)).padStart(2, '0') + ':' +
               String(Math.floor((duree % 3600) / 60)).padStart(2, '0') + ':' +
               String(duree % 60).padStart(2, '0');
    }
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
