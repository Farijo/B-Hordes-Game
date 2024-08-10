const labels = [];
const labelsImg = [];
let imgs = [];
const dataset = [];

function intToColor(num, alpha) {
  const hue = (num * 77) % 360;
  return `hsl(${hue},70%,50%,${alpha})`;
}
function toggleCheckbox(event, el) {
  const input = el.getElementsByTagName('input')[0];
  if(event.target != input) {
    input.checked ^= 1;
    refreshData();
  }
}
function refreshData() {
  if(resizeCanva) {
    $('#myChart').parent().css('min-width', userScale.filter(':checked').length * goalScale.filter(':checked').length * 40)
  }
  myChart.data.datasets = structuredClone(dataset.filter((v,i) => userScale[i].checked));
  myChart.data.labels = labels.filter((v,i) => goalScale[i].checked);
  myChart.data.datasets.forEach(d => d.data = d.data.filter((v,i) => goalScale[i].checked));
  imgs = labelsImg.filter((v,i) => goalScale[i].checked);
  myChart.update();
}
function switchSelection(el) {
  $(el).closest('table').find('input[type=checkbox]').each((i,e) => e.checked ^= 1);
  refreshData();
}
function bindUserLegend() {
    const td = $(document.currentScript.parentElement);
    const idx = dataset.length-1;
    td.css('background-color', dataset[idx].borderColor);
    td.parent().hover(e => {                    
      const d = myChart.data.datasets.find(v => v.label === dataset[idx].label);
      if(d) {
        d.borderWidth = e.type === 'mouseenter' ? 2 : 1;
        myChart.update();
      }
    })
}
function drawChart() {
    const params = new URLSearchParams(window.location.search);
    resizeCanva = params.get('type') === 'bar';
    polarCanva = params.get('type') === 'polarArea';
    myChart = new Chart(document.getElementById('myChart').children[3].children[0], {
      type: params.get('type'),
      plugins: [{
        afterDraw: resizeCanva ? chart => {
          const xAxis = chart.scales.x;
          if(xAxis) {
              const ctx = chart.ctx;
              const yAxis = chart.scales.y;
              xAxis.ticks.forEach((value, index) => {
                if(imgs[index]) {
                  const x = xAxis.getPixelForTick(index);
                  const i = new Image();
                  i.src = "https://myhordes.eu/build/images/"+imgs[index];
                  ctx.drawImage(i, x - 8, yAxis.bottom + 4);
                }
              });
            }
          } : function(chart) {
                const {ctx, data, chartArea: {top, left, right, bottom}, scales: {r}} = chart;
                const centerX = (left + right) / 2;
                const centerY = (top + bottom) / 2;
                const radius = r.drawingArea + 50; // Adjust radius to position icons outside the chart area
    
                data.labels.forEach((label, i) => {
                    if(!imgs[i])return;
                    const angle = r.getIndexAngle(i + (polarCanva ? 0.5:0)) - Math.PI / 2;
                    const x = centerX + Math.cos(angle) * radius;
                    const y = centerY + Math.sin(angle) * radius;
    
                    const imgSize = 16;
                    const im = new Image();
                    im.src = "https://myhordes.eu/build/images/"+imgs[i];
                    ctx.drawImage(im, x - imgSize / 2, y - imgSize / 2, imgSize, imgSize);
                });
            }
      }],
      data: {
        labels: [],
        datasets: []
      },
      options: {
        layout: {
            padding: resizeCanva ? null : {
                top: 50,    // Add padding to top
                bottom: 50, // Add padding to bottom
                left: 50,   // Add padding to left
                right: 50   // Add padding to right
            }
        },
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            display: false
          },
          tooltip: {
            callbacks: {
              title: o => goalScale.filter(':checked')[o[0].dataIndex].parentElement.parentElement.children[1].title
            }
          },
        },
        scales: resizeCanva ? null : {
            r: {
                pointLabels: {
                  display: true,
                  centerPointLabels: polarCanva,
                },
              min: +params.get('min'),
            }
        },
      }
    });
    goalScale = $('#goalScale input[type=checkbox]');
    userScale = $('#userScale input[type=checkbox]');
    $('a.chlg-name[href*='+params.get('type')+']').removeAttr('href', '');
    refreshData();
}
