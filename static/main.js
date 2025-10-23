window.addEventListener("load", () => {
  const notesSectionEl = document.getElementById("notes_section");
  const timerDisplayEl = document.getElementById("timer_display");
  const timerStopEl = document.getElementById("timer_stop");
  const timerStartEl = document.getElementById("timer_start");
  const timerCountdownEl = document.getElementById("timer_countdown");

  const handleUpdate = async () => {
    const paramsString = window.location.search;
    const searchParams = new URLSearchParams(paramsString);
    const client = searchParams.get("client") || "stand_alone";

    const res = await fetch("/data", {
      method: "POST",
      body: JSON.stringify({ poe_client: client }),
    });
    const resJson = await res.json();

    const data = resJson[client];
    if (data) {
      notesSectionEl.innerHTML = data.htmlNote;
    }

    // @todo(nick-ng): change to websocket
    setTimeout(() => {
      handleUpdate();
    }, 2000);
  };

  let timerStartMs = 0;
  let timerCurrentDisplayMs = 0;
  let timerRunning = false;

  const formatTimer = (milliseconds) => {
    let absMilliseconds = Math.abs(milliseconds);
    if (milliseconds < 0) {
      // maths is difficult
      absMilliseconds = absMilliseconds + 1000;
    }
    const seconds = Math.floor((absMilliseconds / 1000) % 60).toString(10)
      .padStart(2, "0");
    const minutes = Math.floor((absMilliseconds / (1000 * 60)) % 60).toString(
      10,
    ).padStart(2, "0");
    const hours = Math.floor(absMilliseconds / (1000 * 60 * 60)).toString(10)
      .padStart(2, "0");

    if (milliseconds < 0) {
      return `-0:00:${seconds}`;
    }

    return [hours, minutes, seconds].join(":");
  };

  const updateTimer = () => {
    if (
      !timerRunning
    ) {
      return;
    }

    timerCurrentDisplayMs = Date.now() - timerStartMs;
    timerDisplayEl.textContent = formatTimer(timerCurrentDisplayMs);

    setTimeout(() => {
      updateTimer();
    }, 100);
  };

  timerStopEl.addEventListener("click", () => {
    timerStartMs = 0;
    timerRunning = false;
  });

  timerStartEl.addEventListener("click", () => {
    if (timerStartMs === 0) {
      timerStartMs = Date.now();
    }

    timerRunning = true;

    const paramsString = window.location.search;
    const searchParams = new URLSearchParams(paramsString);
    const client = searchParams.get("client") || "stand_alone";
    fetch("/reset", {
      method: "POST",
      body: JSON.stringify({ poe_client: client }),
    });

    updateTimer();
  });

  timerCountdownEl.addEventListener("click", () => {
    timerStartMs = Date.now() + 4999;
    timerRunning = true;

    const paramsString = window.location.search;
    const searchParams = new URLSearchParams(paramsString);
    const client = searchParams.get("client") || "stand_alone";
    fetch("/reset", {
      method: "POST",
      body: JSON.stringify({ poe_client: client }),
    });

    updateTimer();
  });

  handleUpdate();
});
