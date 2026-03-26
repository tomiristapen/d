package com.example.nutri_ai_app.water

import java.time.Clock
import java.time.LocalDate
import kotlin.math.roundToInt

class WaterService(
    val userProfile: UserProfile,
    private val activityService: ActivityService,
    private val clock: Clock,
    private val repository: WaterIntakeRepository = InMemoryWaterIntakeRepository(),
) {
    fun calculateWaterGoal(
        weightKg: Double,
        activityLevel: ActivityLevel,
        goal: Goal,
    ): Int {
        require(weightKg > 0) { "weightKg must be > 0" }
        val mlPerKg =
            when (activityLevel) {
                ActivityLevel.LOW -> 30
                ActivityLevel.MODERATE -> 35
                ActivityLevel.HIGH -> 40
            }

        val base = (weightKg * mlPerKg).roundToInt()
        val adjustment =
            when (goal) {
                Goal.WEIGHT_LOSS -> 200
                Goal.MUSCLE_GAIN -> 200
                Goal.MAINTAIN -> 0
                Goal.HEALTHY -> 0
            }
        return base + adjustment
    }

    fun addWater(amountMl: Int) {
        require(amountMl > 0) { "amountMl must be > 0" }
        repository.add(today(), amountMl)
    }

    fun getDailyTotal(): Int = repository.getTotal(today())

    fun calculateProgress(): Double {
        val resolvedActivity = activityService.resolveActivityLevel(userProfile)
        val goalMl =
            calculateWaterGoal(
                weightKg = userProfile.weightKg,
                activityLevel = resolvedActivity,
                goal = userProfile.goal,
            )
        if (goalMl <= 0) return 0.0

        val total = getDailyTotal()
        return (total.toDouble() / goalMl.toDouble()).coerceIn(0.0, 1.0)
    }

    private fun today(): LocalDate = LocalDate.now(clock)
}

