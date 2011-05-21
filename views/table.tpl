%rebase basic
<p>
<a href="/">Return to dashboard</a>
</p>

<p>
service <select id="service_filter">
<option value="" {{ "selected" if selected_service == None else "" }}></option>
% for s in services:
  <option value="{{s.id}}" {{"selected" if selected_service == s else "" }}>{{s.name}}</option>
% end
</select>
</p>
<!-- summary <input type="text">  -->
<input type="button" id="clear-events" value="Clear selected">
<input type="button" id="refresh-events" value="Refresh">
<div id="dynamicdata"></div>
<script type="text/javascript" src="js/table.js"></script>
