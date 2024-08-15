const dataset = [];

function intToColor(num, sat, alpha) {
  const hue = (num * 77) % 360;
  return `hsl(${hue},${20+sat*80}%,50%,${alpha})`;
}
function toggleCheckbox(event, el) {
  const input = el.getElementsByTagName('input')[0];
  if (event.target != input) {
    input.checked ^= 1;
    refreshData();
  }
}
function refreshData() {
  myChart.data.datasets.forEach((d, i) => d.hidden = !userScale[Math.trunc(i/goalScale.length)].checked || !goalScale[i%goalScale.length].checked);
  myChart.update();
}
function switchSelection(el) {
  $(el).closest('table').find('input[type=checkbox]').each((i, e) => e.checked ^= 1);
  refreshData();
}
function bindUserLegend(tdEl, startIdx) {
  $(tdEl).css('background', `linear-gradient(to bottom, ${dataset[startIdx].borderColor} 0%, ${dataset[dataset.length - 1].borderColor} 100%)`);
}
function drawChart() {
  luxon.Settings.defaultLocale = navigator.language
  myChart = new Chart(myChartChildren[0].children[0], {
    type: 'line',
    data: {
      datasets:dataset
    },
    options: {
        scales: {
            x: {
                type: 'time',
                time: {
                    tooltipFormat: 'ff:s,S', // Format détaillé pour la date
                },
            },
        },
        plugins: {
            tooltip: {
                callbacks: {
                  beforeLabel: o => goalScale[o.datasetIndex%goalScale.length].parentElement.parentElement.children[0].title,
                }
            },
            legend: {
                display: false,
            }
        },
        responsive: true,
        maintainAspectRatio: false
    }
});
  refreshData();
}
