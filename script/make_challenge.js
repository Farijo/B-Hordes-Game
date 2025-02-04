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
            if (dataMap[dataItem].name[mh_lang] < opt[ii][1]) {
                opt.splice(ii, 0, [dataItem, dataMap[dataItem].name[mh_lang]]);
                break;
            }
        }
        if (ii == opt.length) {
            opt.push([dataItem, dataMap[dataItem].name[mh_lang]]);
        }
    }
    return opt.reduce((acc, [val, name]) => acc + `<option class="${c}" value="${val}">${name}</option>`, '');
}).join('');

const classes = ['.alp', '.eslc', '.c', '.aeb', '.p'];

addAGoal = function(deletable) {
    const agoal = $(goalhtml);
    const goalTypes = $(agoal[0]);

    const target = agoal.find('#goal-img');
    const selectList = agoal.find('.goal-list');
    const selectValue = agoal.find('#goal-value');
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
        agoal.find(cls).css('display', '').prop('disabled', false);
        selectList.val(selectList.children(cls+":first").val());
        selectList.trigger('change');
    }).trigger('change');

    if(!deletable) {
        goalTypes.children('option[value=5]').remove();
    }
    $('#all-goals').append(agoal);
}

})();

$('#more').on('click', addAGoal);
$('dialog').on('click', e => e.delegateTarget.close());

function bindFormValues(participation, private, goals, api) {
    $('[name=participation]').val(participation).trigger('change');
    $('[name=privat]').val(private ? 1 : 0);
    $('[name=validation_api]').prop('checked', api);

    let deletable = false;

    for (const goal of goals) {
        addAGoal(deletable);
        deletable = true;
        bindGoal(goal);
    }

    if(!deletable) {
        addAGoal(false);
    }
}

function bindGoal(goal) {
    $(`[name=type]:last`).val(goal.Typ).trigger('change');

    switch (goal.Typ) {
        case 1: // case
            $(`[name=x]:last`).val(goal.X.Valid ? goal.X.Byte : "").trigger('change');
            $(`[name=y]:last`).val(goal.Y.Valid ? goal.Y.Byte : "").trigger('change');
        case 0: // picto
        case 3: // banque
            $(`[name=count]:last`).val(goal.Amount.Valid ? goal.Amount.Int32 : "").trigger('change');
        case 2: // construire
            const nval = goal.Entity;
            for (var key in pictImg[goal.Typ]) {
                if (pictImg[goal.Typ][key].id == nval) {
                    $(`.goal-list:last`).val(key).trigger('change');
                    break;
                }
            }
            break;
        case 4: // perso
            const ta = document.createElement('textarea');
            ta.innerHTML = goal.Custom.String;
            $(`[name=custom]:last`).val(goal.Custom.Valid ? ta.value : "").trigger('change');
        default:
            break;
    }
}

function importSpecificGoal(group) {
    $('body').append('<script src=/script/goals_'+group+'.js></script>');
}

let timeout = null;
function exportGoals(btn, txtMain, txtPressed) {
    const out = [];
    for (const [name, value] of new FormData($('#all-goals').closest('form')[0])) {
        switch (name) {
            case "type":
                out.push({Typ: +value});
                break;
            case "x":
                out[out.length-1].X = {Valid: !!value, Byte: +value};
                break;
            case "y":
                out[out.length-1].Y = {Valid: !!value, Byte: +value};
                break;
            case "count":
                out[out.length-1].Amount = {Valid: !!value, Int32: +value};
                break;
            case "val":
                out[out.length-1].Entity = +value;
                break;
            case "custom":
                out[out.length-1].Custom = {Valid: !!value, String: value};
                break;

            default:
                break;
        }
    }
    out[out.length-1].last = true;
    navigator.clipboard.writeText(JSON.stringify(out));
    if (timeout) {
        clearTimeout(timeout);
    }
    btn.innerText = txtPressed;
    timeout = setTimeout(_ => {
        btn.innerText = txtMain;
        timeout = null;
    }, 500);
}

function loadGoals(str) {
    for (const goal of JSON.parse(str)) {
        bindGoal(goal);
        if (!goal.last) {
            addAGoal(true);
        }
    }
}
