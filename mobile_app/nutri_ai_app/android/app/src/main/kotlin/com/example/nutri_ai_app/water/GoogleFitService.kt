package com.example.nutri_ai_app.water

interface GoogleFitService {
    fun hasPermission(): Boolean

    fun readDailySteps(): Long?
}

