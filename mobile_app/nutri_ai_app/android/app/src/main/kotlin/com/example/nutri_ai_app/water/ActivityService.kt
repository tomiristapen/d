package com.example.nutri_ai_app.water

class ActivityService(
    private val googleFitService: GoogleFitService?,
) {
    fun mapStepsToActivity(steps: Long): ActivityLevel {
        require(steps >= 0) { "steps must be >= 0" }
        return when {
            steps < 5_000 -> ActivityLevel.LOW
            steps <= 10_000 -> ActivityLevel.MODERATE
            else -> ActivityLevel.HIGH
        }
    }

    fun resolveActivityLevel(userProfile: UserProfile): ActivityLevel {
        if (userProfile.activitySource != ActivitySource.GOOGLE_FIT) return userProfile.manualActivity
        if (googleFitService == null) return userProfile.manualActivity
        if (!googleFitService.hasPermission()) return userProfile.manualActivity

        val steps = googleFitService.readDailySteps() ?: return userProfile.manualActivity
        return mapStepsToActivity(steps)
    }
}

