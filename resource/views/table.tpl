{{> head}}

<p>
	<a href="/">Return to dashboard</a>
</p>

Filters:
<ul>
{{#filters}}
	<li>
	<form action="/remove-notification-filter" method="POST">
		<input type="submit" value="remove"/>
		<input type="hidden" name="service" value="{{service}}">
		<input type="hidden" name="id" value="{{Id}}"/> 
	</form>
	{{Expression}}
	</li>
{{/filters}}
</ul>

<form action="/add-notification-filter" method="POST">
		<input type="submit" value="Add notification filter" class="span-5">
		<input type="hidden" name="service" value="{{service}}">
		<input type="text" name="regexp" class="span-12 last">
</form>

<hr>

<div class="span-24 last">
<!-- summary <input type="text">  -->
<input type="hidden" id="service_id" value="{{service}}">
<input type="button" id="clear-all-events" value="Clear all">
<input type="button" id="clear-events" value="Clear selected">
<input type="button" id="refresh-events" value="Refresh">
<div id="dynamicdata"></div>
</div>
<script type="text/javascript" src="js/table.js"></script>

{{> foot}}