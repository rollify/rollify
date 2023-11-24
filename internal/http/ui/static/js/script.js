
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