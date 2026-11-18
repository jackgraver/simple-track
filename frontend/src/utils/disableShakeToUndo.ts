// iOS shows a "Undo Typing" alert on device shake whenever a focused text
// field has an undo history, and web apps have no API to turn that off.
// Wiping the undo stack (re-assigning the value) leaves nothing to undo, so
// the alert never appears. Only done on touch-first devices so desktop
// Ctrl/Cmd+Z keeps working.

type TextField = HTMLInputElement | HTMLTextAreaElement

const NON_TEXT_INPUT_TYPES = new Set([
  'checkbox', 'radio', 'range', 'file', 'button', 'submit', 'reset',
  'image', 'color', 'hidden',
])

function isTextField(target: EventTarget | null): target is TextField {
  if (target instanceof HTMLTextAreaElement) return true
  return target instanceof HTMLInputElement && !NON_TEXT_INPUT_TYPES.has(target.type)
}

function clearUndoStack(field: TextField) {
  // Re-assigning the value resets WebKit's undo history.
  const value = field.value
  field.value = ''
  field.value = value
}

export function disableShakeToUndo() {
  if (typeof window === 'undefined' || !window.matchMedia) return
  if (!window.matchMedia('(hover: none) and (pointer: coarse)').matches) return

  // Reset on blur: editing is finished, so caret position doesn't matter.
  document.addEventListener('focusout', (e) => {
    if (isTextField(e.target)) clearUndoStack(e.target)
  })
}
