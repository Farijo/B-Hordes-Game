let addAGoal = null;
const pictImg = [pictos, items, buildings, items];

(function (){

const private = $('[name=privat]').parent().parent();

$('[name=participation]').on('change', function() {
    private.css('display', this.value == 2 ? '':'none');
}).trigger('change');


let selectOpt = [];
for(p in pictos) {
    selectOpt.push(`<option class="alp" value="${p}">${pictos[p].name.fr}</option>`);
}
for(i in items) {
    selectOpt.push(`<option class="eslc aeb" value="${i}">${items[i].name.fr}</option>`);
}
for(b in buildings) {
    selectOpt.push(`<option class="c" value="${b}">${buildings[b].name.fr}</option>`);
}
selectOpt = selectOpt.join('');

const classes = ['.alp', '.eslc', '.c', '.aeb', '.p'];

const goalIndexes = $('[name=goal-indexes]')[0];
let gidx=0;
addAGoal = function(deletable) {
    const agoal = $(goalhtml.replaceAll('{gidx}', ++gidx));
    const goalTypes = $(agoal[0]);

    const target = agoal.children('#goal-img');
    const selectList = agoal.children('#goal-list'+gidx);
    const selectValue = agoal.children('#goal-value');
    selectList.append(selectOpt);
    selectList.on('change', function() {
        if(goalTypes.val() >= pictImg.length) return;
        const ooo = pictImg[goalTypes.val()][this.value];
        target.attr('src', 'https://myhordes.eu/build/images/' + ooo.img);
        selectValue.val(ooo.id);
    });

    const fixedGidx = gidx;
    goalTypes.on('change', function() {
        if(this.value >= classes.length) {
            goalIndexes.value = goalIndexes.value.replace(` ${fixedGidx}`, '');
            agoal.remove();
            return;
        }
        const cls = classes[this.value];
        agoal.find(classes.join(',')).css('display', 'none').prop("disabled", true);
        if(cls != '.p') {
            agoal.find(cls).css('display', '').prop('disabled', false);
            selectList.val(selectList.children(cls+":first").val());
            selectList.trigger('change');
        }
    }).trigger('change');

    if(deletable) {
        goalIndexes.value += ` ${gidx}`;
    } else {
        goalTypes.children('option[value=5]').remove();
    }
    $('#all-goals').append(agoal);
}

})();

addAGoal(null);
$('#more').on('click', addAGoal);

function bindFormValues(participation, private, goals, api) {
    $('[name=participation]').val(participation).trigger('change');
    $('[name=privat]').val(private);
    $('[name=validation_api]').prop('checked', api);

    let i = 0;

    function getNextVal() {
        let val = '';
        while (goals[i] != ':' && i < goals.length) {
            val += goals[i];
            i++;
        }
        i++;
        return val;
    }

    let idx = 0;
    while (i < goals.length) {
        idx++;
        const gvalue = goals[i];
        i += 2;
        $(`[name=type${idx}]`).val(gvalue).trigger('change');

        switch (gvalue) {
            case '1': // case
                $(`[name=x${idx}]`).val(getNextVal()).trigger('change');
                $(`[name=y${idx}]`).val(getNextVal()).trigger('change');
            case '0': // picto
            case '3': // banque
                $(`[name=count${idx}]`).val(getNextVal()).trigger('change');
            case '2': // construire
                const nval = getNextVal();
                for (var key in pictImg[gvalue]) {
                    if (pictImg[gvalue][key].id == nval) {
                        $(`#goal-list${idx}`).val(key).trigger('change');
                        break;
                    }
                  }
                break;
            case '4': // perso
            default:
                i++;
                break;
        }
        if(i >= goals.length) return;
        addAGoal(true);     
    }
}
