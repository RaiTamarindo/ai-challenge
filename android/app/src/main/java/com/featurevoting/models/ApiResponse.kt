package com.featurevoting.models

import com.google.gson.annotations.SerializedName

data class ErrorResponse(
    @SerializedName("error")
    val error: String
)

data class MessageResponse(
    @SerializedName("message")
    val message: String
)