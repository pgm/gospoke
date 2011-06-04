%rebase basic
<a href="/">Return to dashboard</a>
<h1>{{description}}</h1>
<div id="graphdiv" style="width:800px; height:320px;"></div>
<script type="text/javascript" src="/js/graph.js"></script>
<script type="text/javascript">
YAHOO.util.Event.onContentReady("graphdiv", function() {
  var g = new Dygraph(document.getElementById("graphdiv"), "/graph_data/{{name}}");
  });
  
</script>

