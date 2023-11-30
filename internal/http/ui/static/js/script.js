
// cleanDiceSelectors will set empty value (first value) of the selectors
// that are for rolling dices (selector has "diceRollerSelector" class).
function cleanDiceSelectors(){
  let selects = document.querySelectorAll('.diceRollerSelector')
  for (let x of selects) {
    x.value = ""
  }
}

// We will listen for SSE events of new_dice_roll and increase the counter of the badge
// in one.
document.body.addEventListener('htmx:sseMessage', function (evt) {
   // Is this a new dice roll?
  if (evt.detail.type !== "new_dice_roll") {
      return;
  }

  let badge = document.getElementById('notification-badge')
  if (badge == null) {
    return;
  }

  badge.innerHTML = parseInt(badge.innerHTML,10) + 1;
  badge.style.display = 'flex'
});

// Render TS in a prettier ago format.
dayjs.extend(window.dayjs_plugin_relativeTime);
function renderAgoUnixTimestamp(){
   let selected = document.getElementsByClassName("timestamp-ago")
    Array.prototype.forEach.call(selected, element => {
        let ts = dayjs.unix(element.getAttribute("unix-ts"))
        element.innerHTML = ts.fromNow()
        element.setAttribute("data-tooltip", ts.format("YYYY/MM/DD HH:mm")) // Add tooltip using PicoCSS.
    })
}

// Render "Ago" timestamps and update every minute.
renderAgoUnixTimestamp()
setInterval(renderAgoUnixTimestamp, 60000); 

// Every time we update the table with infinite scroll dice rolls using HTMX, update Ago Timestamps. 
document.body.addEventListener('htmx:afterSwap', function (evt) {
  if (evt.detail.elt.id === "history-dice-roll-row") {
      renderAgoUnixTimestamp()
  }
});

