$(document).ready(function() {
  var $customList = $("#custom-list");
  var $boxes = $("#map > div");
  var $selectedBox = null;

  $boxes.on("click", function(event) {
    event.stopPropagation();
  
    var selectedDist = Math.pow(parseInt($(this).data("x")), 2) + Math.pow(parseInt($(this).data("y")), 2);
    selectedDist = Math.round(Math.sqrt(selectedDist));
    var $listItems = $customList.find("li");
    $listItems.each(function() {
      var $listItem = $(this);
      var minDist = parseInt($listItem.data("min"));
      var maxDist = parseInt($listItem.data("max"));
      $listItem.css("display", (minDist <= selectedDist && maxDist >= selectedDist) ? "flex" : "none")
    });

    if($selectedBox) $selectedBox.removeClass('highlight');
    $selectedBox = $(this);

    var html = "Ville";
    if(!$(this).hasClass("city")) {
      $selectedBox.addClass('highlight');
      html = `[${$(this).data("x")},${$(this).data("y")}] ${selectedDist}<span style="font-size:90%">km</span>`;
    }
    $("#info").html(html);
  
  });
  
  $customList.on("click", "li", function(event) {
    var selectedImage = $(this).data("image");
    var selectedText = $(this).find("span").text();
    if ($selectedBox) {
      var $boxImage = $selectedBox.find("img");
      if ($boxImage.length) {
        $boxImage.attr("src", selectedImage);
      } else {
        $selectedBox.append($("<img>").attr("src", selectedImage));
      }
      $selectedBox.attr("title", selectedText);
    }
    $customList.addClass("hidden");
  });
});
