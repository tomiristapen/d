package com.example.nutri_ai_app.water

enum class Goal {
    WEIGHT_LOSS,
    MAINTAIN,
    MUSCLE_GAIN,
    HEALTHY,
}

enum class ActivitySource {
    MANUAL,
    GOOGLE_FIT,
}

enum class ActivityLevel {
    LOW,
    MODERATE,
    HIGH,
}

data class UserProfile(
    val weightKg: Double,
    val goal: Goal,
    val activitySource: ActivitySource,
    val manualActivity: ActivityLevel,
)

