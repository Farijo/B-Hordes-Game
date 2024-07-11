document.addEventListener("DOMContentLoaded", function(event) { 
    var scrollpos = localStorage.getItem('scrollpos');
    if (scrollpos) window.scrollTo(0, scrollpos);
});

window.onbeforeunload = function(e) {
    localStorage.setItem('scrollpos', window.scrollY);
};

const endDateInput = $('[name=end_date]');

function setupDateReactions(dateInput, oldDateTxt, oldDateVal, oldDateConfirm, newDateTxt) {
    dateInput.on('change', function () {
        if(!this.value || new Date(this.value) < new Date()) {
            dateInput.siblings('button').text(oldDateTxt);
            dateInput.siblings('button').val(oldDateVal);
            dateInput.parent().attr('onsubmit', "return confirm(`" + oldDateConfirm + "`)");
        } else {
            dateInput.siblings('button').text(newDateTxt);
            dateInput.siblings('button').val('validate');
            dateInput.parent().attr('onsubmit', null);
        }
    }).trigger('change');
}

function decomposeTemps(duree) {
    return `~ ${Math.floor(duree / 3600).padStart(2, '0')}:${Math.floor((duree % 3600) / 60).padStart(2, '0')}:${(duree % 60).padStart(2, '0')}`;
}

function dateToISOGMT(formData, field) {
    formData.set(field, new Date(formData.get(field)).toISOString().substring(0,16));
}
