window.addEventListener("load", () => {
  const ACT_ZONE_LEVELS = [
    {
      "The Twilight Strand": 1,
      "The Coast": 2,
      "The Tidal Island": 3,
      "The Mud Flats": 4,
      "The Submerged Passage": 5,
      "The Flooded Depths": 6,
      "The Ledge": 6,
      "The Climb": 7,
      "The Lower Prison": 8,
      "The Upper Prison": 9,
      "The Warden's Quarters": 9,
      "Prisoner's Gate": 10,
      "The Ship Graveyard": 11,
      "The Ship Graveyard Cave": 12,
      "The Cavern of Wrath": 12,
      "The Cavern of Anger": 13,
      "Mervil's Lair": 13,
    },
    {},
    {},
    {},
    {},
    {},
    {},
    {},
    {},
    {},
  ];

  const zoneDisplayEl = document.getElementById("zone_display");
  const levelDisplayEl = document.getElementById("level_display");
  const notesSectionEl = document.getElementById("notes_section");
  // const zoneSectionEl = document.getElementById("zone_section");
  // const zoneImageEl = document.getElementById("zone_image");
  // const zoneNotesEl = document.getElementById("zone_notes");
  // const buildSectionEl = document.getElementById("build_section");
  // const buildImageEl = document.getElementById("build_image");
  // const buildNotesEl = document.getElementById("build_notes");

  const allNotes = [];
  let currentIds = "";

  const handleUpdate = async () => {
    let client = "stand_alone";

    const res = await fetch("/data", {
      method: "POST",
      body: JSON.stringify({ poe_client: client }),
    });
    const resJson = await res.json();

    const data = resJson[client];
    if (data) {
      levelDisplayEl.textContent = data.level;
      let zoneLevel = 0;
      for (let i = 0; i < ACT_ZONE_LEVELS.length; i++) {
        const levelsOnly = Object.values(ACT_ZONE_LEVELS[i]);
        const minLevelCutoff = Math.min(...levelsOnly) - 5;
        const maxLevelCutoff = Math.max(...levelsOnly) + 5;
        if (minLevelCutoff <= data.level && data.level <= maxLevelCutoff) {
          ACT_ZONE_LEVELS[i][data.zone];
        }
      }
      if (zoneLevel > 0) {
        zoneDisplayEl.textContent = `${data.zone}, ${zoneLevel}`;
      } else {
        zoneDisplayEl.textContent = data.zone;
      }
    }

    const notes = [];
    allNotes.forEach((note) => {
      if (typeof note.minLevel === "number" && data.level < note.minLevel) {
        return false;
      }

      if (typeof note.maxLevel === "number" && data.level > note.maxLevel) {
        return false;
      }

      if (Array.isArray(note.zones) && !note.zones.includes(data.zone)) {
        return false;
      }

      notes.push(...note.notes);
    });

    const newIds = notes.map((note) => note.id).join(";");
    if (currentIds === newIds) {
      setTimeout(() => {
        handleUpdate();
      }, 500);
      return;
    }

    const newHtml = notes.map((note) => {
      switch (note.type) {
        case "click": {
          return `<code id="${note.id}">${note.text}</code>`;
        }
        default: {
          return `<div id="${note.id}">${note.text}</div>`;
        }
      }
    }).join("");

    notesSectionEl.innerHTML = newHtml;
    if (notes.length > 0) {
      const codeTags = document.querySelectorAll("code");
      codeTags.forEach((codeTag) => {
        if (codeTag.getAttribute("role") !== "button") {
          const text = codeTag.textContent;
          codeTag.addEventListener("click", () => {
            navigator.clipboard.writeText(text);
          });
          codeTag.setAttribute("role", "button");
          codeTag.classList.add("clickToCopy");
        }
      });
    }

    // @todo(nick-ng): change to websocket
    setTimeout(() => {
      handleUpdate();
    }, 2000);
  };

  const loadHelpers = async () => {
    const res = await fetch("/note");
    const resJson = await res.json();

    for (let i = 0; i < resJson.length; i++) {
      const res = await fetch(resJson[i]);
      const resJson2 = await res.json();
      allNotes.push(
        ...resJson2.map((step, ii) => ({
          ...step,
          notes: step.notes.map((note, jj) => {
            const id = `${resJson[i]}_${ii}_${jj}`.replaceAll("/", "_")
              .replaceAll(".json", "");
            if (typeof note === "string") {
              return { id, type: "text", text: note };
            }

            return { ...note, id };
          }),
        })),
      );
    }

    handleUpdate();
  };

  loadHelpers();
});
