var tablerowlink = function(href, target){
  window.open(href, target);
}

$('form#publishreviewform').submit(function(){
  // TODO: validate
  $.post("/review/publish", $(this).serialize(), function(base){
    var data = jQuery.parseJSON(base);
    a = $('#publishreviewview > .alert');
    a.removeClass("alert-success alert-danger");
    if(data.Error) {
      a.addClass("alert-danger");
      a.text("Could not publish code review. Message: " + data.Errormsg);
    } else {
      a.addClass("alert-success");
      a.html('Code Review published! <a href="' + data.CommitURL + '" target="_blank">Take a look at it.<a/>');
    }
    a.show();
  });
  event.preventDefault();
  return false
});

$('a#reviewlisttab').click(function(){
  $('div.result-content').hide();
  $('div#listreviewsview').show();

  // nav active marking
  $('a.list-group-item').removeClass("active")
  $(this).addClass("active")

  $.getJSON("/review/list", {"course": course}, function(data){
    def = $('#reviewlisttable > tbody > tr').last();
    $('#reviewlisttable > tbody').html(def);
    data.Reviews.forEach(function(r, i) {
      if (r.Desc.length > 75) {
        desc = r.Desc.substring(0, 75) + "...";
      } else {
        desc = r.Desc;
      }
      $('#reviewlisttable > tbody > tr').first().before("<tr onclick=\"tablerowlink('" + r.URL + "', '_blank')\"><td>" + (i + 1) + "</td><td>" + r.Title + "</td><td><i>" + desc + "</i></td><td><a href=\"#listcodereview\">Go to review</a></td></tr>");
    });
  });
});

// allow use of tabs in textareas
$("textarea").keydown(function(e) {
  if(e.keyCode === 9) { // tab was pressed
    // get caret position/selection
    var start = this.selectionStart;
    var end = this.selectionEnd;

    var $this = $(this);
    var value = $this.val();

    // set textarea value to: text before caret + tab + text after caret
    $this.val(value.substring(0, start) + "    " + value.substring(end));

    // put caret at right position again (add one for the tab)
    this.selectionStart = this.selectionEnd = start + 4;

    // prevent the focus lose
    e.preventDefault();
  }
});