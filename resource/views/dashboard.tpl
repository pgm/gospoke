{{> head}}

<h2>Services monitored</h2>
<table>
  <tr>
  	<th class="span-1">Status</th>
  	<th class="span-3">Service</th>
  	<th class="span-3">Notifications</th>
	<th class="span-2">Alerts?</th>
	<th class="span-1">Heartbeat</th>
  	<th>Description</th>
  </tr>

{{#services}}
  <tr>
    <td>
      {{#IsDown}}<img src="img/red.png">{{/IsDown}}
      {{#IsUnknown}}<img src="img/gray.png">{{/IsUnknown}}
      {{#IsUp}}<img src="img/green.png">{{/IsUp}}
    </td>
    <td>
      <a href="/list-events?service={{Name}}">{{Name}}</a>
    </td>
    <td>
      {{#Notifications}}
        <span class="event-class-{{Severity}}">{{Count}}</span>
      {{/Notifications}}
    </td>

    <td>
    {{#Enabled}}
      <img src="img/notify_on.png"><div><a href="disable-service?service={{Name}}">Disable</a></div>
    {{/Enabled}}
    {{^Enabled}}
      <a href="enable-service?service={{Name}}">Enable</a>
    {{/Enabled}}
    </td>
    <td>
    {{LastHeartbeatTimestamp}}
    </td>
    <td>
    </td>
  </tr>
{{/services}}

</table>

{{> foot}}
