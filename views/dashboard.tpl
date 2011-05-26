{{> head}}

<h2>Services monitored</h2>
<table>
  <tr>
  	<th class="span-1">Status</th>
  	<th class="span-3">Service</th>
  	<th class="span-3">Notifications</th>
    <th class="span-3">Alerts?</th>
  	<th>Description</th>
  </tr>

{{#services}}
  <tr>
    <td>
      {{#IsDown}}<img src="img/red.png">{{/IsDown}}
      {{#IsUnknown}}<img src="img/red.png">{{/IsUnknown}}
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
      <img src="img/notify_on.png"><a href="disable-service?service={{Name}}">Disable</a>
    {{/Enabled}}
    {{^Enabled}}
      <a href="enable-service?service={{Name}}">Enable</a>
    {{/Enabled}}
    </td>


    <td>
    </td>
  </tr>
{{/services}}

</table>

{{> foot}}
