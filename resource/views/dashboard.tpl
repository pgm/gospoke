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

{{#groups}}
  <tr>
  <td colspan="6">{{Group}}</td>
  </tr>

  {{#Services}}
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
	<div>{{FilterCount}} filters</div>
      {{/Enabled}}
      {{^Enabled}}
        <a href="enable-service?service={{Name}}">Enable</a>
      {{/Enabled}}
      </td>
      <td>
      {{LastHeartbeatTimestamp}}
      </td>
      <td>
      {{Description}}
      </td>
    </tr>
  {{/Services}}
{{/groups}}


</table>

{{> foot}}
