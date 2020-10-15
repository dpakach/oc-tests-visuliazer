function RandomColorNumber() {
  return Math.floor(Math.random() * 256)
}

var MainChart
window.addEventListener('load', () => {
  FetchData("owncloud", "issue")
});

const FetchData = (storage, by) => {
  fetch("/api?storage=" + storage + "&by=" + by)
  .then(data => data.json())
  .then(json => {
    return json
  })
  .then(responseData => {
    responseData = responseData
    console.log(responseData)
    const labels = Object.keys(responseData).map(key => {
      if (by === "issue") {
        return key.replace("https://github.com/owncloud/", "")
      } else {
        return key.split("/")[0]
      }
    })
    var ctx = document.getElementById('graphChart').getContext('2d');
    const backgroundColor = labels.map(() => {
      return `rgba(${RandomColorNumber()}, ${RandomColorNumber()}, ${RandomColorNumber()}, 0.6)`
    })
    MainChart = new Chart(ctx, {
      type: 'horizontalBar',
      data: {
          labels: labels,
          datasets: [{
              label: 'Number of Failing Tests',
              data: Object.values(responseData).map(value => value.length),
              backgroundColor,
              borderWidth: 1
          }]
      },
      options: {
          scales: {
              yAxes: [{
                  ticks: {
                      beginAtZero: true
                  }
              }]
          }
      }
    });
  })
}

document.getElementById("storage-selector").addEventListener("change", (e) => {
  MainChart.destroy()
  let by = document.getElementById("grouping-selector").value
  FetchData(e.target.value, by)
})

document.getElementById("grouping-selector").addEventListener("change", (e) => {
  MainChart.destroy()
  let storage = document.getElementById("storage-selector").value
  FetchData(storage, e.target.value)
})
