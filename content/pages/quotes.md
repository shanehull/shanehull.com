---
title: 🤔 Quotes
description: A randomly generated quote indented to be rarely inspiring, but often realistic. By the way, this page uses htmx and I'm not sorry.
htmx: true
---

Someone ask for a random quote?

Sure! Here's one for you..

<div id="quote-container">
	<blockquote hx-get="/quote" hx-trigger="load"hx-target="#quote-container" hx-swap="swap:0.4s">
        <i></i>
        <footer></footer>
    </blockquote>
</div>
