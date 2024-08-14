const dataset = [];

function intToColor(num, alpha) {
  const hue = (num * 77) % 360;
  return `hsl(${hue},70%,50%,${alpha})`;
}
function toggleCheckbox(event, el) {
  const input = el.getElementsByTagName('input')[0];
  if (event.target != input) {
    input.checked ^= 1;
    refreshData();
  }
}
function refreshData() {
  myChart.data.datasets.forEach((d, i) => d.data.hidden = userScale[i/userScale.length].checked && goalScale[i%goalScale.length].checked);
  myChart.update();
}
function switchSelection(el) {
  $(el).closest('table').find('input[type=checkbox]').each((i, e) => e.checked ^= 1);
  refreshData();
}
function bindUserLegend(tdEl) {
  const td = $(tdEl);
  const idx = dataset.length - 1;
  td.css('background-color', dataset[idx].borderColor);
  td.parent().hover(e => {
    const d = myChart.data.datasets.find(v => v.label === dataset[idx].label);
    if (d) {
      d.borderWidth = e.type === 'mouseenter' ? 2 : 1;
      myChart.update();
    }
  })
}
function drawChart() {
  myChart = new Chart(myChartChildren[0].children[0], {
    type: 'line',
    data: {
      datasets: [{
          label: 'Dataset 1',
          data: [
              {x: '2024-08-01T10:00:00Z', y: 20},
              {x: '2024-08-02T11:00:00Z', y: 30},
              {x: '2024-08-03T12:00:00Z', y: 25},
              {x: '2024-08-04T13:00:00Z', y: 40},
              {x: '2024-08-05T14:00:00Z', y: 35},
          ],
          borderColor: 'rgba(75, 192, 192, 1)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
          fill: false,
      }]
  },
    options: {
        scales: {
            x: {
                type: 'time',
                time: {
                    unit: 'day', // Affiche les jours comme unités
                    tooltipFormat: 'MMM D, YYYY, h:mm a', // Format des dates pour les infobulles
                },
                title: {
                    display: true,
                    text: 'Date'
                }
            },
            y: {
                beginAtZero: true,
                title: {
                    display: true,
                    text: 'Value'
                }
            }
        },
        plugins: {
            legend: {
                display: true,
                position: 'top'
            }
        },
        responsive: true,
        maintainAspectRatio: false
    }});
  refreshData();
}
