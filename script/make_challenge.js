let addAGoal = null;
const pictImg = [pictos, items, buildings, items];

(function (){

const private = $('[name=privat]').parent().parent();

$('[name=participation]').on('change', function() {
    private.css('display', this.value == 2 ? '':'none');
}).trigger('change');


const selectOpt = [[pictos, "alp"], [items, "eslc aeb"], [buildings, "c"]].map(([dataMap, c]) => {
    const opt = [];
    for(dataItem in dataMap) {
        let ii;
        for (ii=0; ii<opt.length; ++ii) {
            if (dataMap[dataItem].name.fr < opt[ii][1]) {
                opt.splice(ii, 0, [dataItem, dataMap[dataItem].name.fr]);
                break;
            }
        }
        if (ii == opt.length) {
            opt.push([dataItem, dataMap[dataItem].name.fr]);
        }
    }
    return opt.reduce((acc, [val, name]) => acc + `<option class="${c}" value="${val}">${name}</option>`, '');
}).join('');

const classes = ['.alp', '.eslc', '.c', '.aeb', '.p'];

addAGoal = function(deletable) {
    const agoal = $(goalhtml);
    const goalTypes = $(agoal[0]);

    const target = agoal.children('#goal-img');
    const selectList = agoal.children('.goal-list');
    const selectValue = agoal.children('#goal-value');
    selectList.append(selectOpt);
    selectList.on('change', function() {
        if(goalTypes.val() >= pictImg.length) return;
        const ooo = pictImg[goalTypes.val()][this.value];
        target.attr('src', 'https://myhordes.eu/build/images/' + ooo.img);
        selectValue.val(ooo.id);
    });

    goalTypes.on('change', function() {
        if(this.value >= classes.length) {
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

    if(!deletable) {
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

    while (i < goals.length) {
        const gvalue = goals[i];
        i += 2;
        $(`[name=type]:last`).val(gvalue).trigger('change');

        switch (gvalue) {
            case '1': // case
                $(`[name=x]:last`).val(getNextVal()).trigger('change');
                $(`[name=y]:last`).val(getNextVal()).trigger('change');
            case '0': // picto
            case '3': // banque
                $(`[name=count]:last`).val(getNextVal()).trigger('change');
            case '2': // construire
                const nval = getNextVal();
                for (var key in pictImg[gvalue]) {
                    if (pictImg[gvalue][key].id == nval) {
                        $(`.goal-list:last`).val(key).trigger('change');
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
