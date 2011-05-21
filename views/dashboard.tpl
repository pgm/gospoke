%rebase basic
<h2>Services monitored</h2>
<table>
  <tr>
  	<th class="span-2"></th>
  	<th class="span-1">Status</th>
  	<th class="span-3">Service</th>
  	<th class="span-3">Notifications</th>
  	<th>Description</th>
  </tr>

% for service in services:
  <tr class="{{'service-enabled' if service.enabled else 'service-disabled'}}">
  	<td class="span-2">
%	if service.enabled:
		<form method="POST" action="disable_service">
			<input type="hidden" name="service_id" value="{{service.service_id}}">
			<input type="submit" value="Disable" name="disable_button">
		</form>
%		else:
		<form method="POST" action="enable_service">
			<input type="hidden" name="service_id" value="{{service.service_id}}">
			<input type="submit" value="Enable" name="enable_button">
		</form>
%		end
	</td>


    <td class="span-1"><img src="img/{{service.color}}.png"></td>
    <td class="span-3"><a href="events?service_id={{service.service_id}}">{{service.name}}</a></td>
    <td class="span-3">
%   for status, count in service.event_counts:
%     if status != None:
        <span class="event-class-{{status}}">{{count}}</span>
%     end
%   end
    </td>
    <td>
    	{{service.description}}
    </td>
% end
</table>

