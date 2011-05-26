{{> head}}

<p>
	<a href="/">Return to dashboard</a>
</p>

<!-- summary <input type="text">  -->
<input type="hidden" id="service_id" value="{{service}}">
<input type="button" id="clear-events" value="Clear selected">
<input type="button" id="refresh-events" value="Refresh">
<div id="dynamicdata"></div>
<script type="text/javascript" src="js/table.js"></script>

{{> foot}}