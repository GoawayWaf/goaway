
loaIpRules = function () {
  var showData = $('#show-data');

  var apiurl = 'http://localhost:8085';
  $.getJSON(apiurl + '/iprule', function (response) {
    var count = 0;
    var header = [];
    var items = response.data.map(function (item) {
      if (typeof item.attributes !== "undefined") {

        var  objAttr = "";

        for(x in item.attributes ){
          if (count === 1) {
            header.push(x)
          }
          if(x === "ttl") {
            item.attributes[x] = item.attributes[x] / 1000000000
          }
          objAttr+= '<td >' + item.attributes[x] + '</td>';
        }
        count++;
        return objAttr
      }
      return ""
    });

    showData.empty();

    if (items.length) {
      var content = '<thead><tr  ><td  >' + header.join('</td><td  >') + '</td></tr></thead>';
      content += '<tr  >' + items.join('</tr><tr >');
      var list = $('<tbody>').html(content);
      var table = $('<table />').html(list);
      showData.append( table   );
    }
  });
  $("#form-iprule").on('submit', function(e){
    e.preventDefault();
    console.log(JSON.stringify($("#form-iprule").serializeObject()));

    $.post(apiurl + '/iprule',
           JSON.stringify($("#form-iprule").serializeObject()), // post data || get data
            function(result) {
               loaIpRules();
              $("#alert-message").html('Created')
            }
    ).fail(function(result){
                    jsonresult = result.responseJSON;
                    if(typeof jsonresult.errors[0] !== "undefined") {
                      $("#alert-message").html(jsonresult.errors[0].title + " " + jsonresult.errors[0].details)
                    } else {
                      $("#alert-message").html(result.responseText)
                    }
                })
  });

  showData.text('Loading Data.');
};

$(document).ready(function () {

  loaIpRules()
});
$.fn.serializeObject = function()
{
  var o = {};
  var a = this.serializeArray();
  $.each(a, function() {
    if (o[this.name]) {
      if (!o[this.name].push) {
        o[this.name] = [o[this.name]];
      }
      o[this.name].push(this.value || '');
    } else {
      o[this.name] = this.value || '';
    }
  });
  return o;
};
